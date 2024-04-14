import { useEffect, useState } from "react";

import Card from "@mui/material/Card";
import { CardContent, Typography } from "@mui/material";

import API from "../api/twc.js";
import UTILS from "../utils/utils.js";

function VehicleStatus() {
  const [vehicle, setVehicle] = useState({
    connected: false,
  });

  useEffect(() => {
    init();
  }, []);

  async function init() {
    let res = await API.getVitals();
    setVehicle({
      connected: res.vehicle_connected,
    });
  }

  return (
    <Card variant="outlined">
      <CardContent>
        <Typography sx={{ fontSize: 24 }} gutterBottom>
          Vehicle Status
        </Typography>
        <Typography>
          Connected: {UTILS.renderBool(vehicle.connected)}
        </Typography>
      </CardContent>
    </Card>
  );
}

export default VehicleStatus;
