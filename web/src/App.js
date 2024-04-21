import { useEffect, useState } from "react";

import Alert from "@mui/material/Alert";
import LoadingButton from "@mui/lab/LoadingButton";
import WifiFindIcon from "@mui/icons-material/WifiFind";

import Grid from "@mui/material/Grid";

import WifiStatus from "./components/WifiStatus";
import VehicleStatus from "./components/VehicleStatus";

import API from "./api/twc.js";
import VitalsChart from "./components/VitalsChart.js";

function App() {
  const [isConnected, setConnected] = useState(false);
  const [isFindingTWC, setIsFindingTWC] = useState(false);
  const [isServerError, setIsServerError] = useState(false);

  useEffect(() => {
    init();
  }, []);

  async function init() {
    setIsFindingTWC(false);

    let res = await API.getInfo();
    setConnected(res.ip && res.ip.length > 0);
  }

  async function findTWC() {
    setIsFindingTWC(true);
    setIsServerError(false);

    await API.find();
    init();
  }

  return (
    <div className="App">
      <Grid container space={2}>
        <Grid item xs={12}>
          <Alert severity="info">
            Tesla Wall Connector Gen 3 is only supported!
          </Alert>
          {isServerError && (
            <Alert severity="warning">Server disconnected!</Alert>
          )}
          {!isConnected && (
            <LoadingButton
              onClick={findTWC}
              loading={isFindingTWC}
              loadingPosition="start"
              startIcon={<WifiFindIcon></WifiFindIcon>}
              variant="outlined"
            >
              Find
            </LoadingButton>
          )}
        </Grid>

        <Grid item xs={6}>
          {isConnected && <WifiStatus></WifiStatus>}
        </Grid>

        <Grid item xs={6}>
          {isConnected && <VehicleStatus></VehicleStatus>}
        </Grid>

        <Grid item xs={12}>
          {isConnected && <VitalsChart></VitalsChart>}
        </Grid>
      </Grid>
    </div>
  );
}

export default App;
