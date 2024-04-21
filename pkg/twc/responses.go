package twc

// Vital structure that represents the vitals from the TWC
type Vital struct {
	ContactorClosed    bool     `json:"contactor_closed"`
	VehicleConnected   bool     `json:"vehicle_connected"`
	Session            int      `json:"session_s"`
	Grid_v             float64  `json:"grid_v"`
	Grid_hz            float64  `json:"grid_hz"`
	VehicleCurrentAmp  float64  `json:"vehicle_current_a"`
	CurrentA_a         float64  `json:"currentA_a"`
	CurrentB_a         float64  `json:"currentB_a"`
	CurrentC_a         float64  `json:"currentC_a"`
	CurrentN_a         float64  `json:"currentN_a"`
	VoltageA_v         float64  `json:"voltageA_v"`
	VoltageB_v         float64  `json:"voltageB_v"`
	VoltageC_v         float64  `json:"voltageC_v"`
	RelayCoil_v        float64  `json:"relay_coil_v"`
	PCBATemp_c         float64  `json:"pcba_temp_c"`
	HandleTemp_c       float64  `json:"handle_temp_c"`
	MCUTemp_c          float64  `json:"mcu_temp_c"`
	Uptime             int      `json:"uptime_s"`
	InputThermopile_uv int      `json:"input_thermopile_uv"`
	Prox_v             float64  `json:"prox_v"`
	PilotHigh_v        float64  `json:"pilot_high_v"`
	PilotLow_v         float64  `json:"pilot_low_v"`
	SessionEnergy_wh   float64  `json:"session_energy_wh"`
	ConfigStatus       int      `json:"config_status"`
	EvseState          int      `json:"evse_state"`
	CurrentAlerts      []string `json:"current_alerts"`
}

// WifiStatus structure that represents the TWC wifi status
type WifiStatus struct {
	SSID           string `json:"wifi_ssid"`
	SignalStrength int    `json:"wifi_signal_strength"`
	RSSI           int    `json:"wifi_rssi"`
	SNR            int    `json:"wifi_snr"`
	Connected      bool   `json:"wifi_connected"`
	IP             string `json:"wifi_infra_ip"`
	Internet       bool   `json:"internet"`
	MAC            string `json:"wifi_mac"`
}

// LifetimeStats structure the represents the TWC lifetime
type LifetimeStats struct {
	ContactorCycles       int     `json:"contactor_cycles"`
	ContactorCyclesLoaded int     `json:"contactor_cycles_loaded"`
	AlertsCount           int     `json:"alerts_count"`
	ThermalFoldbacks      int     `json:"thermal_foldbacks"`
	AvgStartupTemp        float64 `json:"avg_startup_temp"`
	ChargeStarts          int     `json:"charge_starts"`
	Energy_wh             int     `json:"energy_wh"`
	ConnectorCycles       int     `json:"connector_cycles"`
	Uptime                int     `json:"uptime_s"`
	ChargingTime          int     `json:"charging_time_s"`
}

// VitalQueryResponse structure that is used when returning vital stats view REST API
type VitalQueryResponse struct {
	Timestamp string `json:"timestamp"`
	Vital     *Vital `json:"vital"`
}
