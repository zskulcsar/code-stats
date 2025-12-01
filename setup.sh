#!/bin/sh

# *** Functions
envrc() {
    # .envrc for direnv
    echo "Setting up .envrc file"
    cat << ENVRC > .envrc
# Environment variables for codex and contex7
export CODEX_HOME=$(pwd)/.codex
export CONTEXT7_API_KEY=${C7AK}
ENVRC
}

codex_cli() {
    # codex-cli setup
    echo "Setting up .codex/config.toml"
    cat << CONFIG > .codex/config.toml
[projects."$(pwd)"]
trust_level = "trusted"

# Context7 MCP server for coding
[mcp_servers.context7]
args = ["-y", "@upstash/context7-mcp"]
command = "npx"
startup_timeout_sec=30

# Limited network
[sandbox_workspace_write]
network_access = true

CONFIG
}

gitignore() {
    # Patching gitignore for codex and direnv
    echo "Patching .gitignore"
    cat << GI >> .gitignore
# Codex
.codex/

# direnv
.env*

GI
}

# TODO: check for git repository first
add_to_git() {
    # Making sure everything is under git that needs to be under git
    echo "Adding project files to the repo"
    git add --force .codex/skills/*
    git add --force .codex/config.toml
    git add .gitignore
    echo "Cleaning up"
    git add setup.sh
    git add README.md
    git commit -m "Placing project files under version control."
}

clean_up() {
    # Removing leftovers (this is a one-shot command)
    if [ ! -f .testing ]; then
        echo "Deleting self"
        rm -f setup.sh
        rm -f README.md
    fi
}

# *** Main
echo "This script does the setup of the project after cloning. Run before anything."
echo "Mind you! This is a one shot script by default, and deletes itself after run"
echo ""
echo "Provide the CONTEXT7_API_KEY value next:"
read C7AK

envrc
codex_cli
gitignore
add_to_git
clean_up
