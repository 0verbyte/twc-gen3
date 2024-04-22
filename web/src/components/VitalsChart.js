import React, { useEffect, useState } from "react";

import Select from "@mui/material/Select";
import MenuItem from "@mui/material/MenuItem";
import FormControl from "@mui/material/FormControl";
import Box from "@mui/material/Box";
import Grid from "@mui/material/Grid";
import InputLabel from "@mui/material/InputLabel";
import TextField from "@mui/material/TextField";
import Tooltip from "@mui/material/Tooltip";
import InfoIcon from "@mui/icons-material/Info";

import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip as ChartTooltip,
  Legend,
} from "chart.js";
import { Line } from "react-chartjs-2";

import API from "../api/twc.js";
import UTILS from "../utils/utils.js";

ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  ChartTooltip,
  Legend
);

const options = {
  responsive: true,
  plugins: {
    legend: {
      position: "bottom",
    },
    title: {
      display: false,
    },
  },
};

const defaultVital = "vehicle_current_a";
const defaultDuration = "15m";

const vitalKeyValueMap = {
  grid_v: "Grid Voltage",
  vehicle_current_a: "Vehicle Current AMP",
  currentA_a: "Current A",
  currentB_a: "Current B",
  currentC_a: "Current C",
  currentN_a: "Current N",
  voltageA_v: "Voltage A",
  voltageB_v: "Voltage B",
  voltageC_v: "Voltage C",
  relay_coil_v: "Relay Coil Volts",
  pcba_temp_c: "PCBA Temp (C)",
  handle_temp_c: "Handle Temp (C)",
  mcu_temp_c: "MCU Temp (C)",
  uptime_s: "Uptime seconds",
  input_thermopile_uv: "Input Thermopile",
  prox_v: "Prox",
  pilot_high_v: "Pilot High Volts",
  pilot_low_v: "Pilot Low Volts",
  session_energy_wh: "Session Energy",
  config_status: "Config Status",
  evse_state: "Evse State",
};

const onTimeDurationChange = UTILS.doneTyping();

const durationToolTip = () => {
  return (
    <Tooltip title='Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h" (e.g ."300ms", "-1.5h" or "2h45m")'>
      Chart Duration <InfoIcon />
    </Tooltip>
  );
};

function VitalsChart() {
  let [vitalField, setVitalField] = useState(defaultVital);

  let [chartData, setChartData] = useState({ labels: [], datasets: [] });
  let [dataset, setDataset] = useState([]);

  useEffect(() => {
    generateChart(defaultDuration);
  }, []);

  const generateChart = (duration) => {
    API.query(duration)
      .then((data) => {
        let labels = data.map((data) => {
          return data.timestamp;
        });

        setDataset(data);
        setChartData({
          labels: labels,
          datasets: [
            {
              label: vitalKeyValueMap[vitalField],
              data: data.map((d) => d.vital[vitalField]),
              borderColor: "#00cd5e",
              backgroundColor: "#00cd5e",
            },
          ],
        });
      })
      .catch((err) => {
        console.log("chart error: ", err);
      });
  };

  const renderVitalMenuItems = Object.keys(vitalKeyValueMap).map(
    (val, index) => {
      return (
        <MenuItem key={index} value={val}>
          {vitalKeyValueMap[val]}
        </MenuItem>
      );
    }
  );

  const handleVitalFieldChange = (event) => {
    let vitalValue = event.target.value;
    setVitalField(vitalValue);
    setChartData({
      labels: chartData.labels,
      datasets: [
        {
          label: vitalKeyValueMap[vitalValue],
          data: dataset.map((d) => d.vital[vitalValue]),
          borderColor: "#00cd5e",
          backgroundColor: "#00cd5e",
        },
      ],
    });
  };

  return (
    <div>
      <Box sx={{ flexGrow: 1 }}>
        <Grid container spacing={2}>
          <Grid item xs={4}>
            <FormControl fullWidth>
              <InputLabel id="vital">Field</InputLabel>
              <Select
                labelId="vital"
                id="vital"
                value={vitalField}
                label={vitalKeyValueMap[defaultVital]}
                onChange={handleVitalFieldChange}
              >
                {renderVitalMenuItems}
              </Select>
            </FormControl>
          </Grid>
          <Grid item xs={4}>
            <FormControl fullWidth>
              <TextField
                id="timeDuration"
                label={durationToolTip()}
                variant="outlined"
                onKeyUp={(event) => {
                  let duration = event.target.value;
                  onTimeDurationChange(() => {
                    generateChart(duration);
                  }, 1500);
                }}
                defaultValue={defaultDuration}
              />
            </FormControl>
          </Grid>
        </Grid>
      </Box>

      <Line options={options} data={chartData} />
    </div>
  );
}

export default VitalsChart;
