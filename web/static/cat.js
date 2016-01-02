
window.onload = function() {
	console.log("loaded");

	var wsuri = "ws://" + location.host + "/ws"
	sock = new WebSocket(wsuri)
	sock.onopen = function() {
		console.log("Websocket connected to: " + wsuri);
	}
	sock.onclose = function(e) {
		console.log("Websocket connection closed: " + e.code);
	}
	sock.onmessage = function(e) {
		console.log("Websocket message received: " + e.data);

		var output = document.getElementById("console_log");
		var li = document.createElement("li");
		var text = document.createTextNode(e.data)
		li.appendChild(text);
		output.appendChild(li);
	}
	sock.onerror = function(e) {
		console.log("Websocket Error: " + e.data);
	}
}

function startjob(name) {
	var xhttp = new XMLHttpRequest();
	xhttp.open("POST", "/job/"+name+"/start", true);
	xhttp.send();
}
