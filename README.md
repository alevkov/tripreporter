<!--
SPDX-FileCopyrightText: 2023 froggie <legal@frogg.ie>

SPDX-License-Identifier: OSL-3.0
-->

# tripreporter
[![Github Stars](https://img.shields.io/github/stars/effectindex/tripreporter?style=social)](https://github.com/effectindex/tripreporter/stargazers)
[![Github 1.0.0](https://img.shields.io/github/milestones/progress-percent/effectindex/tripreporter/4?color=success)](https://github.com/effectindex/tripreporter/issues)
[![Docker Build](https://img.shields.io/github/actions/workflow/status/effectindex/tripreporter/make.yml?logo=go&logoColor=#00ADD8&branch=master)](https://github.com/effectindex/tripreporter/actions/workflows/make.yml)
[![Docker Build](https://img.shields.io/github/actions/workflow/status/effectindex/tripreporter/docker.yml?logo=docker&logoColor=white&branch=master)](https://github.com/effectindex/tripreporter/actions/workflows/docker.yml)
[![Docker Build](https://img.shields.io/github/actions/workflow/status/effectindex/tripreporter/reuse-compliance.yml?label=reuse%20compliance&branch=master&logo=data:image/svg+xml;base64,PHN2ZyB2ZXJzaW9uPSIxLjEiIHZpZXdCb3g9IjAgMCA1MTIgNTEyIiB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciPgogPHBhdGggZD0ibTI1NiA0OGEyMDggMjA4IDAgMSAxIDAgNDE2IDIwOCAyMDggMCAxIDEgMC00MTZ6bTAgNDY0YTI1NiAyNTYgMCAxIDAgMC01MTIgMjU2IDI1NiAwIDEgMCAwIDUxMnptLTU2LjYtMTk5LjRjLTMxLjItMzEuMi0zMS4yLTgxLjkgMC0xMTMuMXM4MS45LTMxLjIgMTEzLjEgMGM5LjQgOS40IDI0LjYgOS40IDMzLjkgMHM5LjQtMjQuNiAwLTMzLjljLTUwLTUwLTEzMS01MC0xODEgMHMtNTAgMTMxIDAgMTgxIDEzMSA1MCAxODEgMGM5LjQtOS40IDkuNC0yNC42IDAtMzMuOXMtMjQuNi05LjQtMzMuOSAwYy0zMS4yIDMxLjItODEuOSAzMS4yLTExMy4xIDB6IiBmaWxsPSIjZmZmIi8+Cjwvc3ZnPgo=)](https://github.com/effectindex/tripreporter/actions/workflows/reuse-compliance.yml)

## Setting up `.env`

---
⚠️ DO THIS BEFORE FOLLOWING ANY OTHER STEPS BELOW ⚠️

---

Copy `.env.example` to `.env`, and `config/redis.conf.example` to `config/redis.conf`, and (optionally) modify them. 

These files are required for building and running the tripreporter.
This step is necessary for both development and production use.

## Building and running

<details><summary><h3>For production (with Docker)</h3></summary>

This is intended for production use.
This will run on `http://localhost:3000` by default.

```bash
# Normal usage, re-running the command should rebuild quickly by using caches.
docker compose up -d

# If you have issues / want to troubleshoot, use this command to force re-build (add -d to run in background)
docker-compose up --build --force-recreate --no-deps
```

</details>

---

<details><summary><h3>For development (no Docker)</h3></summary>

Before you can run the project outside of Docker, you need to have accessible PostgreSQL and Redis databases running.

1. Setup PostgreSQL

- Either install it via your package manager or follow the instructions on [their official website](https://www.postgresql.org/download/).

- Once installed, start the service. On most systems, this will be `sudo systemctl enable --now postgresql`.

- Now, follow the rest of their official documentation, starting [here](https://www.postgresql.org/docs/current/tutorial-createdb.html), in order to create a database that matches the PostgreSQL variables in `.env`.

- This initial setup is only required once.

2. Setup Redis

- Either install it via your package manager or follow the instructions on [their official website](https://redis.io/download/).

- Once installed, `cd` to the project directory, then run `redis-server --daemonize yes config/redis.conf`.

- Running `redis-server` is required once per boot, whenever you want to be running the project.<br>Optionally, if you would like to enable autostart, edit the `redis-tripreporter.service` file, and then run the following in the project directory:<br>
```bash
mkdir -p ~/.config/systemd/user/
cp config/redis-tripreporter.service ~/.config/systemd/user/
systemctl --user enable --now redis-tripreporter
```

3. Run `make` in the project directory.<br>This is required only once, to have caches of static files for `make dev-server`.

4. Run `make dev-ui` and `make dev-server` in two separate terminals. They should both be running simultaneously. Congratulations, you now have a working development environment, navigate to <http://localhost:3000>.<br>This allows for hot reloading the Vue frontend, just re-run `make dev-server` when making changes to the Go backend.<br><br>The advantage of such a setup means you can replace either the backend or frontend while one of them is running.<br>An example of a senario where this can be incredibly useful is when testing different versions of the code at a certain commit to find out when a bug was introduced, to assist with debugging.<br>Another useful scenario is if you want to make a clean `git clone` of the project and use either the backend / frontend from it, while running the "dirty" local counterpart to rule out places in the codebase where a bug could come from, or if you'd like to swap between two versions of the frontend to quickly compare the changes made to it.

</details>
