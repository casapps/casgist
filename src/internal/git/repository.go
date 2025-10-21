package git

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/filesystem"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/google/uuid"
)

// Repository represents a Git repository for a gist
type Repository struct {
	gistID     uuid.UUID
	path       string
	repository *git.Repository
}

// RepositoryManager handles Git operations for gists
type RepositoryManager struct {
	basePath   string
	authorName string
	authorEmail string
}

// NewRepositoryManager creates a new Git repository manager
func NewRepositoryManager(basePath, authorName, authorEmail string) *RepositoryManager {
	return &RepositoryManager{
		basePath:    basePath,
		authorName:  authorName,
		authorEmail: authorEmail,
	}
}

// CreateRepository creates a new Git repository for a gist
func (rm *RepositoryManager) CreateRepository(gistID uuid.UUID) (*Repository, error) {
	repoPath := filepath.Join(rm.basePath, gistID.String())
	
	// Ensure directory exists
	if err := os.MkdirAll(repoPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create repository directory: %w", err)
	}
	
	// Initialize bare repository
	fs := osfs.New(repoPath)
	storer := filesystem.NewStorage(fs, nil)
	
	repo, err := git.Init(storer, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize git repository: %w", err)
	}
	
	return &Repository{
		gistID:     gistID,
		path:       repoPath,
		repository: repo,
	}, nil
}

// OpenRepository opens an existing Git repository
func (rm *RepositoryManager) OpenRepository(gistID uuid.UUID) (*Repository, error) {
	repoPath := filepath.Join(rm.basePath, gistID.String())
	
	// Check if repository exists
	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("repository does not exist: %s", repoPath)
	}
	
	// Open repository
	fs := osfs.New(repoPath)
	storer := filesystem.NewStorage(fs, nil)
	
	repo, err := git.Open(storer, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to open git repository: %w", err)
	}
	
	return &Repository{
		gistID:     gistID,
		path:       repoPath,
		repository: repo,
	}, nil
}

// CommitFiles commits files to the repository
func (r *Repository) CommitFiles(files map[string]string, message string, authorName, authorEmail string) (string, error) {
	// Get worktree
	worktree, err := r.repository.Worktree()
	if err != nil {
		return "", fmt.Errorf("failed to get worktree: %w", err)
	}
	
	// Write files to worktree
	for filename, content := range files {
		filePath := filepath.Join(worktree.Filesystem.Root(), filename)
		
		// Ensure directory exists
		if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			return "", fmt.Errorf("failed to create file directory: %w", err)
		}
		
		// Write file
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			return "", fmt.Errorf("failed to write file %s: %w", filename, err)
		}
		
		// Add to index
		if _, err := worktree.Add(filename); err != nil {
			return "", fmt.Errorf("failed to add file %s to index: %w", filename, err)
		}
	}
	
	// Create commit
	commit, err := worktree.Commit(message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  authorName,
			Email: authorEmail,
			When:  time.Now(),
		},
	})
	
	if err != nil {
		return "", fmt.Errorf("failed to commit: %w", err)
	}
	
	return commit.String(), nil
}

// GetFiles retrieves all files from the latest commit
func (r *Repository) GetFiles() (map[string]string, error) {
	// Get HEAD reference
	head, err := r.repository.Head()
	if err != nil {
		if err == plumbing.ErrReferenceNotFound {
			// No commits yet
			return map[string]string{}, nil
		}
		return nil, fmt.Errorf("failed to get HEAD: %w", err)
	}
	
	// Get commit object
	commit, err := r.repository.CommitObject(head.Hash())
	if err != nil {
		return nil, fmt.Errorf("failed to get commit object: %w", err)
	}
	
	// Get tree
	tree, err := commit.Tree()
	if err != nil {
		return nil, fmt.Errorf("failed to get tree: %w", err)
	}
	
	files := make(map[string]string)
	
	// Walk through tree
	err = tree.Files().ForEach(func(file *object.File) error {
		content, err := file.Contents()
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", file.Name, err)
		}
		files[file.Name] = content
		return nil
	})
	
	if err != nil {
		return nil, fmt.Errorf("failed to walk tree: %w", err)
	}
	
	return files, nil
}

// GetFile retrieves a specific file from the latest commit
func (r *Repository) GetFile(filename string) (string, error) {
	files, err := r.GetFiles()
	if err != nil {
		return "", err
	}
	
	content, exists := files[filename]
	if !exists {
		return "", fmt.Errorf("file not found: %s", filename)
	}
	
	return content, nil
}

// GetHistory retrieves commit history
func (r *Repository) GetHistory(limit int) ([]*Commit, error) {
	// Get HEAD reference
	head, err := r.repository.Head()
	if err != nil {
		if err == plumbing.ErrReferenceNotFound {
			// No commits yet
			return []*Commit{}, nil
		}
		return nil, fmt.Errorf("failed to get HEAD: %w", err)
	}
	
	// Get commit iterator
	commits, err := r.repository.Log(&git.LogOptions{
		From:  head.Hash(),
		Order: git.LogOrderCommitterTime,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get log: %w", err)
	}
	
	var history []*Commit
	count := 0
	
	err = commits.ForEach(func(commit *object.Commit) error {
		if limit > 0 && count >= limit {
			return fmt.Errorf("limit reached") // Use error to break the loop
		}
		
		history = append(history, &Commit{
			Hash:      commit.Hash.String(),
			Message:   commit.Message,
			Author:    commit.Author.Name,
			Email:     commit.Author.Email,
			Timestamp: commit.Author.When,
		})
		
		count++
		return nil
	})
	
	// Ignore limit reached error
	if err != nil && !strings.Contains(err.Error(), "limit reached") {
		return nil, fmt.Errorf("failed to iterate commits: %w", err)
	}
	
	return history, nil
}

// CreateBranch creates a new branch
func (r *Repository) CreateBranch(branchName string) error {
	// Get HEAD reference
	head, err := r.repository.Head()
	if err != nil {
		return fmt.Errorf("failed to get HEAD: %w", err)
	}
	
	// Create new branch reference
	refName := plumbing.NewBranchReferenceName(branchName)
	ref := plumbing.NewHashReference(refName, head.Hash())
	
	if err := r.repository.Storer.SetReference(ref); err != nil {
		return fmt.Errorf("failed to create branch: %w", err)
	}
	
	return nil
}

// ListBranches lists all branches
func (r *Repository) ListBranches() ([]string, error) {
	refs, err := r.repository.References()
	if err != nil {
		return nil, fmt.Errorf("failed to get references: %w", err)
	}
	
	var branches []string
	
	err = refs.ForEach(func(ref *plumbing.Reference) error {
		if ref.Name().IsBranch() {
			branches = append(branches, ref.Name().Short())
		}
		return nil
	})
	
	if err != nil {
		return nil, fmt.Errorf("failed to iterate references: %w", err)
	}
	
	return branches, nil
}

// Clone creates a clone of this repository
func (r *Repository) Clone(targetPath string) error {
	// Create target directory
	if err := os.MkdirAll(targetPath, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}
	
	// Clone repository
	_, err := git.PlainClone(targetPath, false, &git.CloneOptions{
		URL: r.path,
	})
	
	if err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}
	
	return nil
}

// GetSize calculates repository size
func (r *Repository) GetSize() (int64, error) {
	var size int64
	
	err := filepath.WalkDir(r.path, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		
		if !d.IsDir() {
			info, err := d.Info()
			if err != nil {
				return err
			}
			size += info.Size()
		}
		
		return nil
	})
	
	if err != nil {
		return 0, fmt.Errorf("failed to calculate repository size: %w", err)
	}
	
	return size, nil
}

// Delete removes the repository
func (r *Repository) Delete() error {
	return os.RemoveAll(r.path)
}

// Path returns the repository path
func (r *Repository) Path() string {
	return r.path
}

// ID returns the gist ID
func (r *Repository) ID() uuid.UUID {
	return r.gistID
}

// Commit represents a Git commit
type Commit struct {
	Hash      string
	Message   string
	Author    string
	Email     string
	Timestamp time.Time
}