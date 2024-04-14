let serverURL = "";
let debug = false;

// When running the application with `npm start` the server URL needs to be set explicitly since the go backend
// server is running on a different port.
if (window.location.port === "3000") {
  serverURL = "http://localhost:8080";
  debug = true;
}

async function request(path) {
  const response = await fetch(serverURL + path);
  const results = await response.json();
  return results;
}

var TWC = {
  getInfo: async () => {
    if (debug) {
      console.log("API: getInfo()");
    }
    return request("/api/v1/info");
  },

  getWifiStatus: async () => {
    if (debug) {
      console.log("API: getWifiStatus()");
    }
    return request("/api/v1/wifi_status");
  },

  getVitals: async () => {
    if (debug) {
      console.log("API: getVitals()");
    }
    return request("/api/v1/vitals");
  },

  getLifetime: async () => {
    if (debug) {
      console.log("API: getLifetime()");
    }
    return request("/api/v1/lifetime");
  },

  find: async () => {
    if (debug) {
      console.log("API: find()");
    }
    return request("/api/v1/find");
  },
};

export default TWC;
