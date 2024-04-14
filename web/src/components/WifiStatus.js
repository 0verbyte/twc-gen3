import { useEffect, useState } from "react";

import Card from "@mui/material/Card";
import { CardContent, Typography } from "@mui/material";

import API from "../api/twc.js";
import UTILS from "../utils/utils.js";

function WifiStatus() {
  const [wifi, setWifi] = useState({
    ssid: "",
    signalStrength: 0,
    rssi: 0,
    snr: 0,
    connected: false,
    ip: "",
    internet: false,
    mac: "",
  });

  useEffect(() => {
    init();
  }, []);

  async function init() {
    let res = await API.getWifiStatus();
    setWifi({
      ssid: atob(res.wifi_ssid),
      signalStrength: res.wifi_signal_strength,
      rssi: res.wifi_rssi,
      snr: res.wifi_snr,
      connected: res.wifi_connected,
      ip: res.wifi_infra_ip,
      internet: res.internet,
      mac: res.wifi_mac,
    });
  }

  return (
    <Card variant="outlined">
      <CardContent>
        <Typography sx={{ fontSize: 24 }} gutterBottom>
          Wifi Status
        </Typography>
        <Typography>Name: {wifi.ssid}</Typography>
        <Typography>Connected: {UTILS.renderBool(wifi.connected)}</Typography>
        <Typography>IP: {wifi.ip}</Typography>
        <Typography>Internet: {UTILS.renderBool(wifi.internet)}</Typography>
        <Typography>Mac Address: {wifi.mac}</Typography>
      </CardContent>
    </Card>
  );
}

export default WifiStatus;
