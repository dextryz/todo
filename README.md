# TODO

A todo list management system leveraging the [nostr](www.nostr.com) protocol.

```shell
export NOSTR_TODO=$HOME/.config/nostr/todo.json
```

```shell
> nt add food Steak
Creating new list: food

> nt add food Butter

> nt list food
[ ] Steak
[ ] Butter

> nt done food Butter

> nt list food
[ ] Steak
[X] Butter
```
