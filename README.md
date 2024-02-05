# TODO

A todo list management system leveraging the [nostr](www.nostr.com) protocol.

## Setup

Create a config file at `~/.config/nostr/todo.json`

```
{
    "relays": ["wss://relay.damus.io/"],
    "nsec": "nsec..."
}
```

Add environment variable to point to configuration

```shell
export NOSTR_TODO=$HOME/.config/nostr/todo.json
```

Build the binay

```shell
make build
```

## Usage

```shell
> todo add food Steak
Creating new list: food

> todo add food Butter

> todo list food
b1f19350-86d9-4caa-8989-660dcd98df55 (2024-02-05): [ ] Steak
57de7e02-8a1e-4d38-bd6e-53b28ed6e876 (2024-02-05): [ ] Butter

> todo done food b1f19350-86d9-4caa-8989-660dcd98df55

> todo list food
b1f19350-86d9-4caa-8989-660dcd98df55 (2024-02-05): [X] Steak
57de7e02-8a1e-4d38-bd6e-53b28ed6e876 (2024-02-05): [ ] Butter

> todo done food b1f19350-86d9-4caa-8989-660dcd98df55

> todo list food
b1f19350-86d9-4caa-8989-660dcd98df55 (2024-02-05): [ ] Steak
57de7e02-8a1e-4d38-bd6e-53b28ed6e876 (2024-02-05): [ ] Butter
```
