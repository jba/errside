# Errors as Side Notes

If Java-style catch blocks are like footnotes in a book, then Go error-handling
statements are like parentheticals. Both are annoying. The first takes you too
far away from the main text, while the second is so close to it that it's
distracting.

Side notes are a great alternative: the extra material is just an eye-shift
away, but you can easily ignore it if you want to.

This package displays Go error-handling on the right side of the screen.

Before:
```
func (ds *GitDataSource) Fetch(from, to string) ([]string, error) {
	fmt.Printf("Fetching data from %s into %s...\n", from, to)
	if err := createFolderIfNotExist(to); err != nil {
		return nil, err
	}
	if err := clearFolder(to); err != nil {
		return nil, err
	}
	if err := cloneRepo(to, from); err != nil {
		return nil, err
	}
	dirs, err := getContentFolders(to)
	if err != nil {
		return nil, err
	}
	fmt.Println("Fetching complete.")
	return dirs, nil
}
```

After:
```
func (ds *GitDataSource) Fetch(from, to string) ([]string, error) {
    fmt.Printf("Fetching data from %s into %s...\n", from, to)
    createFolderIfNotExist(to)                      =: err; if err != nil { return nil, err }
    clearFolder(to)                                 =: err; if err != nil { return nil, err }
    cloneRepo(to, from)                             =: err; if err != nil { return nil, err }
    dirs := getContentFolders(to)                   =: err; if err != nil { return nil, err }
    fmt.Println("Fetching complete.")
    return dirs, nil
}
```


Note: the following packages were copied from the go/ subtree of the standard
library:
- ast
- importer
- internal/*
- parser
- printer
- types

