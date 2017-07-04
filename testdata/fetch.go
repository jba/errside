package example

import "fmt"

type GitDataSource int

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

func createFolderIfNotExist(string) error        { return nil }
func clearFolder(string) error                   { return nil }
func cloneRepo(string, string) error             { return nil }
func getContentFolders(string) ([]string, error) { return nil, nil }
