export enum ServerEvent {
  LocationMutation = "location.mutation",
  ItemMutation = "item.mutation",
  LabelMutation = "tag.mutation",
}

export type EventMessage = {
  event: ServerEvent;
};

let socket: WebSocket | null = null;

const listeners = new Map<ServerEvent, (() => void)[]>();

function connect(onmessage: (m: EventMessage) => void) {
  let protocol = "ws";
  if (window.location.protocol === "https:") {
    protocol = "wss";
  }

  const dev = import.meta.dev;

  const host = dev ? window.location.host.replace("3000", "7745") : window.location.host;

  const ws = new WebSocket(`${protocol}://${host}/api/v1/ws/events`);

  ws.onopen = () => {
    console.debug("connected to server");
  };

  ws.onclose = () => {
    console.debug("disconnected from server");
    setTimeout(() => {
      connect(onmessage);
    }, 3000);
  };

  ws.onerror = err => {
    console.error("websocket error", err);
  };

  const thorttled = new Map<ServerEvent, (m: EventMessage) => void>();

  thorttled.set(ServerEvent.LocationMutation, useThrottleFn(onmessage, 1000));
  thorttled.set(ServerEvent.ItemMutation, useThrottleFn(onmessage, 1000));
  thorttled.set(ServerEvent.LabelMutation, useThrottleFn(onmessage, 1000));

  ws.onmessage = msg => {
    const pm = JSON.parse(msg.data);
    const fn = thorttled.get(pm.event);
    if (fn) {
      fn(pm);
    }
  };

  socket = ws;
}

export function onServerEvent(event: ServerEvent, callback: () => void) {
  if (socket === null) {
    connect(e => {
      console.debug("received event", e);
      listeners.get(e.event)?.forEach(c => c());
    });
  }

  onMounted(() => {
    if (!listeners.has(event)) {
      listeners.set(event, []);
    }
    listeners.get(event)?.push(callback);
  });

  onUnmounted(() => {
    const got = listeners.get(event);
    if (got) {
      listeners.set(
        event,
        got.filter(c => c !== callback)
      );
    }

    if (listeners.get(event)?.length === 0) {
      listeners.delete(event);
    }
  });
}
