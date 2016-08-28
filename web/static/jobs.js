
window.onload = function() {
	console.log("loaded");

	var wsuri = "ws://" + location.host + "/jobs/ws"
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
			case "addJob":
				var jobs = document.getElementById("jobs");
				var li = document.createElement("li");
				li.id = msg.data.name;
				a = document.createElement("a");
				a.href = "job/" + msg.data.name;
				a.innerHTML = msg.data.name;
				li.appendChild(a);
				li.appendChild(document.createElement("br"));
				li.appendChild(document.createTextNode("Origin Repository: " + msg.data.repo));
				li.appendChild(document.createElement("br"));
				var start = document.createElement("button");
				start.type = "button";
				start.innerHTML = "Start Job";
				start.onclick = function(){ startJob(msg.data.name); };
				li.appendChild(start);
				var del = document.createElement("button");
				del.type = "button";
				del.onclick = function(){ deleteJob(msg.data.name); };
				del.innerHTML = "Delete Job";
				li.appendChild(del);
				jobs.appendChild(li);
				break;
			case "deleteJob":
				var jobs = document.getElementById("jobs");
				var li = document.getElementById(msg.data.name);
				jobs.removeChild(li);
				break;
		}

	}
	sock.onerror = function(e) {
		console.log("Websocket Error: " + e.data);
	}
}
