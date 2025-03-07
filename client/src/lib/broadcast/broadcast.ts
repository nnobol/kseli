let channel: BroadcastChannel | null = null;
type RoomCallback = (roomId: string | null, error: string | null) => void;
let callbacks: RoomCallback[] = [];

function ensureChannel() {
    if (!channel) {
        channel = new BroadcastChannel("active-room");
        channel.onmessage = (event) => {
            const { roomId, error } = event.data;
            callbacks.forEach((cb) => cb(roomId, error));
        };
    }
    return channel;
}


export function broadcastRoomInfo(roomId: string | null, error: string | null) {
    const ch = ensureChannel();
    ch.postMessage({ roomId, error });
}

export function onRoomBroadcast(callback: RoomCallback) {
    ensureChannel();
    callbacks.push(callback);
    return () => {
        callbacks = callbacks.filter((cb) => cb !== callback);
        if (callbacks.length === 0 && channel) {
            channel.close();
            channel = null;
        }
    };
}