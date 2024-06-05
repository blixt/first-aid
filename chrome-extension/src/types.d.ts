interface NetworkInformation extends EventTarget {
	readonly type:
		| "bluetooth"
		| "cellular"
		| "ethernet"
		| "mixed"
		| "none"
		| "other"
		| "unknown"
		| "wifi";
}

interface Navigator {
	readonly connection: NetworkInformation;
}
