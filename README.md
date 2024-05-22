# Panthor-Update-Notification

## Table of Contents

- [Panthor-Update-Notification](#panthor-update-notification)
  - [Table of Contents](#table-of-contents)
  - [About](#about)
    - [Features](#features)
  - [Getting started](#getting-started)

## About

This is an update notification programm for the "Panthor Mod". It regularly checks for new versions of Panthor, and sends notifications to specified webhooks when a new version is available. This ensures that you are always informed about the latest updates without needing to manually check for new versions.

### Features

<details>
<summary>Trigger webhook when a new version is avaiable</summary>

Define what urls should be triggered when a Panthor mod version is avaiable.
Webhooks will receive this payload:

```json
{
  "content": "New version 2.0.5.2 is available!",
  "hasModUpdate": "false",
  "releaseAt": "2024-05-23 00:00:00",
  "size": "",
  "version": "2.0.5.2"
}
```

</details>

## Getting started

1. Clone the repository

   ```bash
   git clone git@github.com:tklein1801/panthor-update-notification.git
   ```

2. Install dependencies

   ```bash
   go mod tidy
   ```

3. Configure `config.yml`

   ```yml
   app:
     interval: '1 * * * *'
     load_on_startup: true
   notification:
     webhooks:
       - https://webhook.site/THIS_IS_SOMETHING
   ```

4. Run the programm

   ```bash
   go run main.go
   ```

5. Programm bauen

   ```bash
   go build
   ```
