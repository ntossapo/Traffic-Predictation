package google

type GoogleRouteRequest struct {
	GeocodedWaypoints []struct {
		GeocoderStatus string `json:"geocoder_status"`
		PlaceID string `json:"place_id"`
		Types []string `json:"types"`
	} `json:"geocoded_waypoints"`
	Routes []struct {
		Bounds struct {
			       Northeast struct {
						 Lat float64 `json:"lat"`
						 Lng float64 `json:"lng"`
					 } `json:"northeast"`
			       Southwest struct {
						 Lat float64 `json:"lat"`
						 Lng float64 `json:"lng"`
					 } `json:"southwest"`
		       } `json:"bounds"`
		Copyrights string `json:"copyrights"`
		Legs []struct {
			Distance struct {
					 Text string `json:"text"`
					 Value int `json:"value"`
				 } `json:"distance"`
			Duration struct {
					 Text string `json:"text"`
					 Value int `json:"value"`
				 } `json:"duration"`
			EndAddress string `json:"end_address"`
			EndLocation struct {
					 Lat float64 `json:"lat"`
					 Lng float64 `json:"lng"`
				 } `json:"end_location"`
			StartAddress string `json:"start_address"`
			StartLocation struct {
					 Lat float64 `json:"lat"`
					 Lng float64 `json:"lng"`
				 } `json:"start_location"`
			Steps []struct {
				Distance struct {
						 Text string `json:"text"`
						 Value int `json:"value"`
					 } `json:"distance"`
				Duration struct {
						 Text string `json:"text"`
						 Value int `json:"value"`
					 } `json:"duration"`
				EndLocation struct {
						 Lat float64 `json:"lat"`
						 Lng float64 `json:"lng"`
					 } `json:"end_location"`
				HTMLInstructions string `json:"html_instructions"`
				Polyline struct {
						 Points string `json:"points"`
					 } `json:"polyline"`
				StartLocation struct {
						 Lat float64 `json:"lat"`
						 Lng float64 `json:"lng"`
					 } `json:"start_location"`
				TravelMode string `json:"travel_mode"`
				Maneuver string `json:"maneuver,omitempty"`
			} `json:"steps"`
			TrafficSpeedEntry []interface{} `json:"traffic_speed_entry"`
			ViaWaypoint []interface{} `json:"via_waypoint"`
		} `json:"legs"`
		OverviewPolyline struct {
			       Points string `json:"points"`
		       } `json:"overview_polyline"`
		Summary string `json:"summary"`
		Warnings []interface{} `json:"warnings"`
		WaypointOrder []interface{} `json:"waypoint_order"`
	} `json:"routes"`
	Status string `json:"status"`
}
