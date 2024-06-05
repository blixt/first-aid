const WEBSOCKET_URL = "ws://localhost:49158/control";
let activeWebSocket: WebSocket | undefined;

interface RpcCall {
	id: number;
	method: string;
	// biome-ignore lint/suspicious/noExplicitAny: <explanation>
	params: any;
}

interface RpcResult {
	id: number;
	result?: unknown;
	error?: string;
}

interface TabDetails {
	id: number;
	url: string;
	title: string;
}

function getErrorMessage(error: unknown, fallback = "Unknown error"): string {
	if (error instanceof Error) {
		return error.message;
	}
	return String(error) || fallback;
}

function connectWebSocket() {
	if (activeWebSocket && activeWebSocket.readyState !== WebSocket.CLOSED) {
		return;
	}

	const socket = new WebSocket(WEBSOCKET_URL);
	activeWebSocket = socket;

	socket.onopen = () => {
		console.log("WebSocket connected");
	};

	socket.onclose = (event) => {
		console.log("WebSocket closed, reconnecting...", event);
		setTimeout(connectWebSocket, 5000); // Attempt to reconnect every 5 seconds
	};

	socket.onerror = (error) => {
		console.log("WebSocket error", error);
		socket.close();
	};

	socket.onmessage = async (event) => {
		console.log("WebSocket message received:", event.data);
		let rpcCall: RpcCall | undefined;
		try {
			rpcCall = JSON.parse(event.data) as RpcCall;
			if (!rpcCall) throw Error("Invalid RPC call");
			const { id, method, params } = rpcCall;
			let result: unknown;

			switch (method) {
				case "getTabs":
					result = await getTabs();
					break;
				case "setActiveTab":
					result = await setActiveTab(params);
					break;
				case "openTab":
					result = await openTab(params.urlOrQuery, params.background);
					break;
				case "screenshotTab":
					result = await screenshotTab(params.id);
					break;
				default:
					throw new Error(`Unknown method: ${method}`);
			}

			const rpcResult: RpcResult = { id, result };
			socket.send(JSON.stringify(rpcResult));
		} catch (error) {
			console.error("Error handling WebSocket message:", error);
			if (!rpcCall) {
				return;
			}
			const response: RpcResult = {
				id: rpcCall.id,
				error: getErrorMessage(error),
			};
			socket.send(JSON.stringify(response));
		}
	};
}

async function getTabs(): Promise<TabDetails[]> {
	try {
		const tabs = await chrome.tabs.query({});
		const tabDetails = [];
		for (let i = 0; i < tabs.length; i++) {
			const tab = tabs[i];
			if (tab.id === undefined) continue;
			tabDetails.push({
				id: tab.id,
				url: tab.url || "",
				title: tab.title || "",
			});
		}
		return tabDetails;
	} catch (error) {
		throw new Error(getErrorMessage(error, "Failed to get tabs"));
	}
}

async function setActiveTab(id: number): Promise<void> {
	try {
		await chrome.tabs.update(id, { active: true });
	} catch (error) {
		throw new Error(getErrorMessage(error, "Failed to set active tab"));
	}
}

async function openTab(
	urlOrQuery: string,
	background: boolean,
): Promise<number> {
	try {
		const tab = await chrome.tabs.create({
			url: urlOrQuery,
			active: !background,
		});
		if (tab.id === undefined) {
			throw new Error("Failed to create tab");
		}
		return tab.id;
	} catch (error) {
		throw new Error(getErrorMessage(error, "Failed to open tab"));
	}
}

async function screenshotTab(id: number): Promise<string> {
	try {
		// Activate the provided tab id.
		await chrome.tabs.update(id, { active: true });
		// Capture the visible area of the currently active tab.
		const dataURL = await chrome.tabs.captureVisibleTab({ format: "png" });
		return dataURL;
	} catch (error) {
		throw new Error(getErrorMessage(error, "Failed to capture screenshot"));
	}
}
// Connect when the service worker is activated
self.addEventListener("activate", () => {
	console.log("Service worker activated");
	connectWebSocket();
});

// Reconnect on network status change
navigator.connection.addEventListener("change", () => {
	if (navigator.onLine) {
		console.log("Network status changed: online");
		connectWebSocket();
	} else {
		console.log("Network status changed: offline");
	}
});

self.addEventListener("fetch", () => {
	// This empty fetch listener ensures the service worker stays active
});

// Reconnect when the service worker is started
self.addEventListener("startup", () => {
	console.log("Service worker started");
	connectWebSocket();
});
