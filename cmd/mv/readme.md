# File mover

This is not really complete yet, but it makes it easier to move files between packages in a large go project.

To use it:

1. Edit the main.go file to change the file to be renamed
2. `go run ./cmd/mv/main.go`
3. Note the random hex value which is appended to each of the filenames, this will be a key for renaming
4. Move the file
5. Do a local find-and-replace in the file
  * Replacing `([a-zA-Z0-9_]*)_<key>` with `$1`
6. Do a global find-and-replace
  * For files in the package where you're removing the file from
  * Excluding sub-folders
  * Replacing:  `([a-zA-Z0-9_]*)_<key>` with `<new package>.$1`
7. Do a global find-and-replace
  * For files in any package
  * Replacing:  `<old package>.([a-zA-Z0-9_]*)_<key>` with `<new package>.$1`
8. Do a global search for `<key>`, if you find it then something went wrong

**NOTE**: Using VSCode to do the global find-and-replace is nice because it will automatically add the
imports of the new package where necessary, otherwise you'll need to make a list of all files containing
`<key>` and add an import of the new package there.