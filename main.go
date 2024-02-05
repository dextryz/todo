package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip19"
)

var ErrNotFound = errors.New("todo list not found")

func StringEnv(key string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		log.Fatalf("env variable \"%s\" not set, usual", key)
	}
	return value
}

var CONFIG = StringEnv("NOSTR_TODO")

type Config struct {
	Nsec   string   `json:"nsec"`
	Relays []string `json:"relays"`
}

type Todo struct {
	Id        string `json:"id"`
	Content   string `json:"content"`
	Done      bool   `json:"done"`
	CreatedAt int64  `json:"created_at"`
}

type TodoList []Todo

func loadConfig() (*Config, error) {
	b, err := os.ReadFile(CONFIG)
	if err != nil {
		return nil, err
	}
	var cfg Config
	err = json.Unmarshal(b, &cfg)
	if err != nil {
		return nil, err
	}
	if len(cfg.Relays) == 0 {
		log.Println("please set atleast on relay in your config.json")
	}
	return &cfg, nil
}

func tagName(name string) string {
	if name == "" {
		return "nostr-todo"
	}
	return "nostr-todo-" + name
}

func (tl *TodoList) MarshalJSON() ([]byte, error) {
	return json.Marshal(*tl)
}

func (tl *TodoList) UnmarshalJSON(b []byte) error {
	return json.Unmarshal(b, (*[]Todo)(tl))
}

func (tl *TodoList) Sort() {
	sort.Slice(*tl, func(i, j int) bool {
		return (*tl)[i].CreatedAt < (*tl)[j].CreatedAt
	})
}

func (tl *TodoList) Load(ctx context.Context, cfg *Config, name string) error {

	pool := nostr.NewSimplePool(ctx)
	filter := nostr.Filter{
		Kinds: []int{nostr.KindApplicationSpecificData},
		Tags: nostr.TagMap{
			"d": []string{tagName(name)},
		},
	}

	ev := pool.QuerySingle(ctx, cfg.Relays, filter)
	if ev == nil {
		return ErrNotFound
	}

	return tl.UnmarshalJSON([]byte(ev.Content))
}

func (tl *TodoList) Save(ctx context.Context, cfg *Config, name string) error {

	var sk string
	var pub string
	if _, s, err := nip19.Decode(cfg.Nsec); err == nil {
		sk = s.(string)
		if pub, err = nostr.GetPublicKey(s.(string)); err != nil {
			return err
		}
	} else {
		return err
	}

	b, err := tl.MarshalJSON()
	if err != nil {
		return err
	}

	e := nostr.Event{
		Kind:      nostr.KindApplicationSpecificData,
		Content:   string(b),
		CreatedAt: nostr.Now(),
		PubKey:    pub,
		Tags: nostr.Tags{
			{"d", tagName(name)},
		},
	}

	err = e.Sign(sk)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	for _, relayURL := range cfg.Relays {
		wg.Add(1)

		relayURL := relayURL
		go func() {
			defer wg.Done()

			relay, err := nostr.RelayConnect(ctx, relayURL)
			if err != nil {
				log.Println(err)
				return
			}
			defer relay.Close()

            err = relay.Publish(ctx, e)
			if err != nil {
				log.Println(err)
				return
			}
		}()
	}
	wg.Wait()

	return nil
}

func add(ctx context.Context, cfg *Config, name, content string) error {

	var tl TodoList
	err := tl.Load(ctx, cfg, name)

	if err == ErrNotFound {
		fmt.Printf("- Creating new list: %s\n", name)
	} else if err != nil {
		return err
	}

	tl = append(tl, Todo{
		Id:        uuid.New().String(),
		Content:   content,
		Done:      false,
		CreatedAt: time.Now().Unix(),
	})

	tl.Sort()

	return tl.Save(ctx, cfg, name)
}

func list(ctx context.Context, cfg *Config, name string) error {

	var tl TodoList
	err := tl.Load(ctx, cfg, name)
	if err != nil {
		return err
	}

	for _, todo := range tl {
		mark := "[ ]"
		if todo.Done {
			mark = "[X]"
		}
		fmt.Printf("%s (%s): %s %s\n",
			todo.Id,
			time.Unix(todo.CreatedAt, 0).Format("2006-01-02"),
			mark,
			todo.Content,
		)
	}

	return nil
}

func done(ctx context.Context, cfg *Config, name, id string) error {

	var tl TodoList
	err := tl.Load(ctx, cfg, name)
	if err != nil {
		return err
	}

	for i := 0; i < len(tl); i++ {
		if tl[i].Id != id {
			continue
		}
		tl[i].Done = true
	}

	return tl.Save(ctx, cfg, name)
}

func undone(ctx context.Context, cfg *Config, name, id string) error {

	var tl TodoList
	err := tl.Load(ctx, cfg, name)
	if err != nil {
		return err
	}

	for i := 0; i < len(tl); i++ {
		if tl[i].Id != id {
			continue
		}
		tl[i].Done = false
	}

	return tl.Save(ctx, cfg, name)
}

func main() {

	flag.Usage = func() {
		fmt.Printf("Usage: %s [-lines | -bytes] [files...]\n\n", os.Args[0])
		fmt.Println("Counts words (or lines or bytes) in named files or standard input.\n\nFlags:")
		flag.PrintDefaults()
	}

	flag.Parse()

	cfg, err := loadConfig()
	if err != nil {
		log.Fatalln(err)
	}

	args := os.Args[1:]
	ctx := context.Background()

	if args[0] == "list" {
		err := list(ctx, cfg, args[1])
		if err != nil {
			fmt.Printf("ERROR: %v\n", err)
		}
	}

	if args[0] == "add" {
		err := add(ctx, cfg, args[1], args[2])
		if err != nil {
			fmt.Printf("ERROR: %v\n", err)
		}
	}

	if args[0] == "done" {
		err := done(ctx, cfg, args[1], args[2])
		if err != nil {
			fmt.Printf("ERROR: %v\n", err)
		}
	}

	if args[0] == "undone" {
		err := undone(ctx, cfg, args[1], args[2])
		if err != nil {
			fmt.Printf("ERROR: %v\n", err)
		}
	}
}
