GOALS
1. Develop a LRU cache with Get/Set API
2. Build a react app that consumes the LRU cache api

REQUIREMENTS
Develop a LRU Cache
The cache will store Key/Value with expiration. If the expiration for key is set to 5 seconds,
then that key should be evicted from the cache after 5 seconds. The cache can store maximum of
1024 keys.

Must Haves
● Backend should be built on Golang
● The Get/set method in cache should be exposed as api endpoints

Good to have
● Implementing concurrency in cache

Develop a React Application
● Develop a react application that will consume Get api to get the key from cache and set
key/value in the cache

RUN :-
client -> npm start
server -> go run main.go
