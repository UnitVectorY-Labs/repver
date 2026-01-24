---
layout: default
title: repver
nav_order: 1
permalink: /
---

# repver

Keep version numbers and other shared strings consistent across your repo, without hunting through files by hand.

`repver` applies a set of pre-defined replacements across one or more files, and can optionally handle the Git workflow to deliver the change as a pull request.

## What it does

You describe an update once in a `.repver` file:
- which files to touch
- which line patterns to match
- which named values you want to replace
- what Git operations to perform (branch, commit, push, PR)

Then you run one simple command to apply that update consistently, reducing your toil and risk of human error.

## When it’s useful

If a value appears in more than one place, `repver` helps you change it safely and repeatably, for example:

- Updating language or runtime versions across build files and CI workflows (Go, Node.js, Python, etc.)
- Keeping Dockerfiles, docs, and release notes in sync with the version you actually ship
- Standardizing updates across many repos that follow the same conventions

## How it complements Dependabot

Dependabot is great at updating dependencies it understands.

`repver` is for the other category of version strings: the ones you chose to embed in places like documentation, CI configuration, Dockerfiles, or custom metadata. It’s not a replacement for Dependabot, it’s a practical add-on for the gaps.
