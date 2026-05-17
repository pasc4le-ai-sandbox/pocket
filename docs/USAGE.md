# Usage

## Adding items

Add one or more files/directories to the pocket clipboard. Globbing is supported.

```bash
pocket report.pdf
pocket images/ notes.txt
pocket *.log
```

Each call **appends** to the current clipboard. The clipboard is not cleared
until you release or delete items from it.

## Listing items

Show the current clipboard contents, one per line with a 1-indexed number:

```bash
pocket --list
pocket -l
```

Example output:

```
1  /home/user/report.pdf
2  /home/user/images/
3  /home/user/notes.txt
```

## Removing items

Remove an item from the clipboard by its list number. **This does not delete
the underlying file** — it only removes the reference from the clipboard.

```bash
pocket --delete 2
pocket -d 2
```

## Releasing items

Copy all clipboard items to the **current working directory**:

```bash
pocket --release
pocket -r
```

If a destination file already exists, the item is skipped with a warning but
other items continue to be processed. On full success the clipboard is cleared.

## Moving items (cut)

Use `--cut` together with `--release` to **move** files instead of copying
them:

```bash
pocket --release --cut
pocket -r -c
```

This renames (moves) each item to the current directory. Directories are
moved whole.

## Combined example

```bash
# Gather files
pocket ~/Downloads/report.pdf ~/Pictures/screenshot.png ~/Documents/project/

# Oops, don't need the screenshot — remove it
pocket -l
# 1  /home/user/Downloads/report.pdf
# 2  /home/user/Pictures/screenshot.png
# 3  /home/user/Documents/project/
pocket -d 2

# Add another
pocket ~/notes.txt

# Move everything to current directory
pocket -r -c
```
