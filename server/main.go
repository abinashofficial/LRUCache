package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/cors"
)

type CacheEntry struct {
	Key        string      `json:"key"`
	Value      interface{} `json:"value"`
	Expiration time.Time   `json:"expiration"`
}

type LRUCache struct {
	capacity  int
	cache     map[string]CacheEntry
	order     []string
	mutex     sync.Mutex
	clients   map[*websocket.Conn]bool
	broadcast chan map[string]interface{}
}

func NewLRUCache(capacity int) *LRUCache {
	return &LRUCache{
		capacity:  capacity,
		cache:     make(map[string]CacheEntry),
		order:     make([]string, 0),
		clients:   make(map[*websocket.Conn]bool),
		broadcast: make(chan map[string]interface{}),
	}
}

func (c *LRUCache) Get(key string) (interface{}, bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	entry, ok := c.cache[key]
	if !ok {
		return nil, false
	}

	// Check if the entry has expired
	if time.Now().After(entry.Expiration) {
		delete(c.cache, key)
		c.removeFromOrder(key)
		return nil, false
	}

	c.moveToFront(key)
	return entry.Value, true
}

func (c *LRUCache) Set(key string, value interface{}, expiration time.Time) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// If key exists, update its value and expiration
	if _, ok := c.cache[key]; ok {
		c.cache[key] = CacheEntry{Key: key, Value: value, Expiration: expiration}
		c.moveToFront(key)
		return
	}

	// Evict the least recently used entry if the cache is full
	if len(c.cache) >= c.capacity {
		delete(c.cache, c.order[len(c.order)-1])
		c.order = c.order[:len(c.order)-1]
	}

	// Add new entry
	c.cache[key] = CacheEntry{Key: key, Value: value, Expiration: expiration}
	c.order = append([]string{key}, c.order...)
	c.broadcastCacheData()
}

func (c *LRUCache) Delete(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	delete(c.cache, key)
	c.removeFromOrder(key)
	c.broadcastCacheData()
}

func (c *LRUCache) moveToFront(key string) {
	for i, k := range c.order {
		if k == key {
			copy(c.order[1:], c.order[:i])
			c.order[0] = key
			return
		}
	}
}

func (c *LRUCache) removeFromOrder(key string) {
	for i, k := range c.order {
		if k == key {
			c.order = append(c.order[:i], c.order[i+1:]...)
			return
		}
	}
}

func (c *LRUCache) broadcastCacheData() {
	cacheData := make(map[string]interface{})
	for key, entry := range c.cache {
		cacheData[key] = entry
	}
	for client := range c.clients {
		err := client.WriteJSON(cacheData)
		if err != nil {
			log.Printf("error: %v", err)
			client.Close()
			delete(c.clients, client)
		}
	}
}

func main() {
	cache := NewLRUCache(100)

	// WebSocket endpoint for cache updates
	http.HandleFunc("/cacheUpdates", func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Upgrade(w, r, nil, 1024, 1024)
		if err != nil {
			log.Println(err)
			return
		}
		defer conn.Close()

		cache.clients[conn] = true

		// Read from WebSocket (keep alive)
		for {
			if _, _, err := conn.NextReader(); err != nil {
				break
			}
		}
	})

	// REST API endpoints
	http.HandleFunc("/get", func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		value, ok := cache.Get(key)
		if !ok {
			http.Error(w, "Key not found", http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(value)
	})

	http.HandleFunc("/set", func(w http.ResponseWriter, r *http.Request) {
		var data struct {
			Key   string      `json:"key"`
			Value interface{} `json:"value"`
			TTL   int         `json:"ttl"`
		}
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		expiration := time.Now().Add(time.Duration(data.TTL) * time.Second)
		cache.Set(data.Key, data.Value, expiration)
		w.WriteHeader(http.StatusCreated)
	})

	http.HandleFunc("/delete", func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		cache.Delete(key)
		w.WriteHeader(http.StatusNoContent)
	})

	handler := cors.Default().Handler(http.DefaultServeMux)

	fmt.Println("Server started at localhost:8080")
	log.Fatal(http.ListenAndServe(":8081", handler))
}

// package main

// import (
//   "database/sql"
//   "fmt"
//   "log"

//   _ "github.com/lib/pq"
// )
// var db *sql.DB

// type User struct {
//     ID    int
//     Name  string
// }

// func getUser(id int) {
//     var user User
//     sqlStatement := `SELECT id, name FROM studentinfo WHERE id=$1`
//     row := db.QueryRow(sqlStatement, id)
//     switch err := row.Scan(&user.ID, &user.Name); err {
//     case sql.ErrNoRows:
//         fmt.Println("No rows were returned!")
//     case nil:
//         fmt.Printf("Fetched single record: %v\n", user)
//     default:
//         log.Fatalf("Unable to scan the row. %v", err)
//     }
// }

// func main() {
//   connStr := "postgresql://develop_owner:fkdK1b9vzohQ@ep-green-feather-a1lerkc8.ap-southeast-1.aws.neon.tech/develop?sslmode=require"
//   db, err := sql.Open("postgres", connStr)
//   if err != nil {
//     log.Fatal(err)
//   }
//   defer db.Close()

// //   rows, err := db.Query("SELECT * FROM studentinfo")
// //   if err != nil {
// //     log.Fatal(err)
// //   }
// //   defer rows.Close()

// //   var version string
// //   for rows.Next() {
// //     err := rows.Scan(&version)
// //     if err != nil {
// //       log.Fatal(err)
// //     }
// //   }
// var user User
// sqlStatement := `SELECT id, name FROM studentinfo WHERE id=$1`
// row := db.QueryRow(sqlStatement, 101)
// switch err := row.Scan(&user.ID, &user.Name); err {
// case sql.ErrNoRows:
// 	fmt.Println("No rows were returned!")
// case nil:
// 	fmt.Printf("Fetched single record: %v\n", user)
// default:
// 	log.Fatalf("Unable to scan the row. %v", err)
// }
//   fmt.Printf("version=%s\n", row)
// }
