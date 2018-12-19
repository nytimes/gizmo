package service

import (
	"html/template"
	"net/http"

	"github.com/NYTimes/gizmo/server"
)

// Demo will serve an HTML page that demonstrates how to use the 'stream'
// endpoint.
func (s *StreamService) Demo(w http.ResponseWriter, r *http.Request) {
	vals := struct {
		Port     int
		StreamID int64
	}{
		s.port,
		server.GetInt64Var(r, "stream_id"),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := demoTempl.Execute(w, &vals)
	if err != nil {
		server.Log.Error("template error ", err)
		http.Error(w, "problems loading HTML", http.StatusInternalServerError)
	}
}

var demoTempl = template.Must(template.New("demo").Parse(demoHTML))

const demoHTML = `<!DOCTYPE html>
<html lang="en">

<head>
<title>StreamService Demo</title>
</head>

<body>
	<h1>Welcome to the stream for {{ .StreamID }}!</h1>
	<p>Open multiple tabs to see messages broadcast across all views</p>
	<div id="consumed" style="float:left; width:50%">
	</div>
	<div id="published" style="float:left">
	</div>
	<script src="https://ajax.googleapis.com/ajax/libs/jquery/2.1.3/jquery.min.js"></script>
	<script type="text/javascript">
		(function()
		{
			var conn = new WebSocket(
				"ws://localhost:{{ .Port }}/svc/v1/stream/{{ .StreamID }}"
			);

			// consume from websocket/Kafka
			conn.onmessage = function(evt)
			{
				var evts = $("#consumed");
				evts.prepend("<p> Received: " + evt.data + "</p>");
			}

			// publish to websocket/Kafka
			setInterval(publishMessage(conn), 1000);

			function publishMessage(conn) {
				return function() {
					var alpha = "ABCDEFGHIJKLMNOPQRSTUVWXYZ";
					var msg = '{"game":"crossword","user_id":12345,"time":'+new Date().getTime()+',"cell":' + Math.floor(
						(Math.random() * 10) + 1) +
						',"value":"' + alpha.charAt(Math.floor(
							Math.random() * alpha.length)) + '"}'
							conn.send(msg);

						var evts = $("#published");
						evts.prepend("<p> Sent: " + msg + "</p>");
					}
				}
			})();
	</script>
</body>
</html>`
