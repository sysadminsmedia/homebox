import { useViewPreferences } from "./use-preferences";
import { watch } from "vue";

export enum ServerEvent {
  EntityMutation = "entity.mutation",
  TagMutation = "tag.mutation",
  UserMutation = "user.mutation",
  ExportMutation = "export.mutation",
  ImportMutation = "import.mutation",
}

export type EventMessage = {
  event: ServerEvent;
};

let socket: WebSocket | null = null;
let currentTenantId: string | null = null;
let watcherSetup = false;
let reconnectTimer: ReturnType<typeof setTimeout> | null = null;
let reconnectAttempts = 0;

const RECONNECT_BASE_DELAY_MS = 1000;
const RECONNECT_MAX_DELAY_MS = 30000;

const listeners = new Map<ServerEvent, (() => void)[]>();

function getWebSocketProtocols() {
  const auth = useAuthContext();
  if (!auth.attachmentToken) {
    return undefined;
  }

  // Browser WebSocket APIs cannot set arbitrary headers, so pass auth in the
  // subprotocol header and parse it server-side.
  return ["hb-auth", auth.attachmentToken];
}

function clearReconnectTimer() {
  if (reconnectTimer !== null) {
    clearTimeout(reconnectTimer);
    reconnectTimer = null;
  }
}

function nextReconnectDelay() {
  const delay = Math.min(RECONNECT_BASE_DELAY_MS * 2 ** reconnectAttempts, RECONNECT_MAX_DELAY_MS);
  reconnectAttempts += 1;
  return delay;
}

function scheduleReconnect(onmessage: (m: EventMessage) => void) {
  if (reconnectTimer !== null) {
    return;
  }

  const delay = nextReconnectDelay();
  reconnectTimer = setTimeout(() => {
    reconnectTimer = null;
    if (!useAuthContext().attachmentToken) {
      return;
    }
    connect(onmessage);
  }, delay);
}

function connect(onmessage: (m: EventMessage) => void) {
  let protocol = "ws";
  if (window.location.protocol === "https:") {
    protocol = "wss";
  }

  const dev = import.meta.dev;

  const host = dev ? window.location.host.replace("3000", "7745") : window.location.host;

  let url = `${protocol}://${host}/api/v1/ws/events`;
  if (currentTenantId) {
    url += `?tenant=${currentTenantId}`;
  }

  const protocols = getWebSocketProtocols();
  if (!protocols) {
    return;
  }

  const ws = new WebSocket(url, protocols);

  ws.onopen = () => {
    reconnectAttempts = 0;
    clearReconnectTimer();
    console.debug("connected to server");
  };

  ws.onclose = () => {
    console.debug("disconnected from server");
    socket = null;
    scheduleReconnect(onmessage);
  };

  ws.onerror = err => {
    console.error("websocket error", err);
  };

  const thorttled = new Map<ServerEvent, (m: EventMessage) => void>();

  thorttled.set(ServerEvent.EntityMutation, useThrottleFn(onmessage, 1000));
  thorttled.set(ServerEvent.TagMutation, useThrottleFn(onmessage, 1000));
  thorttled.set(ServerEvent.UserMutation, useThrottleFn(onmessage, 1000));
  thorttled.set(ServerEvent.ExportMutation, useThrottleFn(onmessage, 500));
  thorttled.set(ServerEvent.ImportMutation, useThrottleFn(onmessage, 500));

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
  const prefs = useViewPreferences();
  currentTenantId = prefs.value.collectionId || null;

  if (!watcherSetup) {
    watch(
      () => prefs.value.collectionId,
      newId => {
        currentTenantId = newId || null;
        reconnectAttempts = 0;
        clearReconnectTimer();

        if (socket) {
          socket.onclose = null;
          socket.close();
          socket = null;
        }

        connect(e => {
          console.debug("received event", e);
          listeners.get(e.event)?.forEach(c => c());
        });
      }
    );
    watcherSetup = true;
  }

  if (socket === null) {
    reconnectAttempts = 0;
    clearReconnectTimer();
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
