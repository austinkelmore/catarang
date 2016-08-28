
window.onload = function() {
	console.log("loaded");

	var wsuri = "ws://" + location.host + location.pathname + "/ws"
	sock = new WebSocket(wsuri)
	sock.onopen = function() {
		console.log("Websocket connected to: " + wsuri);
	}
	sock.onclose = function(e) {
		console.log("Websocket connection closed: " + e.code);
	}
	sock.onmessage = function(e) {
		console.log("Websocket message received: " + e.data);

		var msg = JSON.parse(e.data);
		console.log(msg);

		switch(msg.type) {
			case "consoleLog":
				var output = document.getElementById("console_log");
				var li = document.createElement("li");
				var text = document.createTextNode(msg.data);
				li.appendChild(text);
				output.appendChild(li);
				break;
		}

	}
	sock.onerror = function(e) {
		console.log("Websocket Error: " + e.data);
	}
}
