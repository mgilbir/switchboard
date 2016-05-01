package switchboard

import (
	"bufio"
	"fmt"
	"net/http"
	"strings"
	"sync"
)

var (
	EmptyBlacklist = NewBlacklist("")
)

type Blacklist struct {
	domains  []string
	category string
	lock     sync.RWMutex
}

func NewBlacklist(category string) Blacklist {
	return Blacklist{
		domains:  make([]string, 0),
		category: category,
	}
}

func (b *Blacklist) Add(domain string) {
	b.lock.Lock()
	defer b.lock.Unlock()
	b.domains = append(b.domains, domain)
}

func (b *Blacklist) Domains() []string {
	b.lock.RLock()
	defer b.lock.RUnlock()

	d := make([]string, len(b.domains))
	for i, v := range b.domains {
		d[i] = v
	}
	return d
}

func (b Blacklist) Category() string {
	return b.category
}

func RetrieveBlacklist(src string, category string) (Blacklist, error) {
	if strings.HasPrefix(src, "http") {
		return RetrieveBlacklistURL(src, category)
	}

	return EmptyBlacklist, fmt.Errorf("Unknown blacklist source type for: %s\n", src)
}

func RetrieveBlacklistURL(url string, category string) (Blacklist, error) {
	r, err := http.Get(url)
	if err != nil {
		return EmptyBlacklist, err
	}

	defer r.Body.Close()

	bl := NewBlacklist(category)

	scanner := bufio.NewScanner(r.Body)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		isComment := strings.HasPrefix(line, "#")

		if len(line) > 0 && !isComment {
			bl.Add(line)
		}
	}

	return bl, nil
}
