import React, { useState, useEffect } from "react";
import axios from "axios";

interface CacheItem {
  key: string;
  value: any;
  expiration: string; // Assuming expiration is a string for display purposes
}

const App: React.FC = () => {
  const [key, setKey] = useState("");
  const [getValue, setGetValue] = useState("");
  const [getKey, setGetKey] = useState("");


  const [value, setValue] = useState("");
  const [ttl, setTTL] = useState(5); // Default TTL of 5 seconds
  const [cacheData, setCacheData] = useState<{ [key: string]: CacheItem }>({});

  // Establish WebSocket connection
  useEffect(() => {
    const ws = new WebSocket("ws://localhost:8081/cacheUpdates");
    ws.onopen = () => {
      console.log("WebSocket connected");
    };
    ws.onmessage = (event) => {
      const data = JSON.parse(event.data);
      setCacheData(data);
    };
    ws.onclose = () => {
      console.log("WebSocket closed");
    };
    return () => {
      ws.close();
    };
  }, []);

  const handleGet = async () => {
    try {
      const response = await axios.get(`http://localhost:8081/get?key=${getKey}`);
      console.log(response.data);
      setGetValue(response.data)
    } catch (error) {
      console.error("Error fetching key:", error);
      setGetValue("No Data")
    }
  };

  const handleSet = async () => {
    try {
      await axios.post("http://localhost:8081/set", {
        key: key,
        value: value,
        ttl: ttl,
      });
    } catch (error) {
      console.error("Error setting key-value:", error);
    }
  };

  const handleDelete = async () => {
    try {
      await axios.get(`http://localhost:8081/delete?key=${key}`);
    } catch (error) {
      console.error("Error deleting key:", error);
    }
  };
  return (
    <div>
      <h1>LRU Cache React App</h1>
      <div>
        <label>Key:</label>
        <input type="text" value={key} onChange={(e) => setKey(e.target.value)} />
      </div>
      <div>
        <label>Value:</label>
        <input type="text" value={value} onChange={(e) => setValue(e.target.value)} />
      </div>
      <div>
        <label>TTL (seconds):</label>
        <input type="number" value={ttl} onChange={(e) => setTTL(parseInt(e.target.value))} />
      </div>
      <div>
        {/* <button onClick={handleGet}>Get Key</button> */}
        <button onClick={handleSet}>Set Key-Value</button>
        <button onClick={handleDelete}>Delete Key</button>

      </div>
      <div>
        <h2>Cache Data:</h2>
        <ul>
          {Object.entries(cacheData).map(([key, entry]) => (
            <li key={key}>
              {key}: {entry.value}, Expiration: {entry.expiration}
            </li>
          ))}
        </ul>
      </div>
      <div>

      <h2>Get Data:</h2>

      <div>
        <label>Key:</label>
        <input type="text" value={getKey} onChange={(e) => setGetKey(e.target.value)} />
        <button onClick={handleGet}>Get Key</button>

      </div>
      {/* <ul>
          {Object.entries(cacheData).map(([key, entry]) => (
            <li key={key}>
              {key}: {entry.value}, Expiration: {entry.expiration}
            </li>
          ))}
        </ul> */}
                <label>{getValue}</label>

      </div>

    </div>
  );
};

export default App;
