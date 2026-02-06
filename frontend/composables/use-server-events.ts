import { useViewPreferences } from "./use-preferences";
import { watch } from "vue";
import { Centrifuge, type TransportEndpoint } from "centrifuge";

export enum ServerEvent {
  LocationMutation = "location.mutation",
  ItemMutation = "item.mutation",
  TagMutation = "tag.mutation",
}

export type EventMessage = {
  event: ServerEvent;
};

let centrifuge: Centrifuge | null = null;
let currentTenantId: string | null = null;
let watcherSetup = false;

const listeners = new Map<ServerEvent, (() => void)[]>();

async function getConnectionToken(): Promise<string> {
  const prefs = useViewPreferences();

  const headers: Record<string, string> = {
    "Content-Type": "application/json",
  };

  if (prefs?.value?.collectionId) {
    headers["X-Tenant"] = prefs.value.collectionId;
  }

  console.debug("Fetching token with tenant:", prefs?.value?.collectionId || "none");

  const response = await fetch("/api/v1/ws/token", {
    method: "GET",
    credentials: "include",
    headers,
  });

  console.debug("Token response status:", response.status);

  if (!response.ok) {
    const errorText = await response.text();
    console.error("Failed to get connection token:", response.status, errorText);
    console.error("Request headers:", headers);
    throw new Error(`Failed to get connection token: ${response.status} - ${errorText}`);
  }

  const data = await response.json();
  console.debug("Received connection token:", data.token ? "✓" : "✗", "length:", data.token?.length || 0);
  return data.token;
}

async function connect(onmessage: (m: EventMessage) => void) {
  const dev = import.meta.dev;

  // In dev mode, use the backend port directly
  // In prod, use the same host as the frontend
  const host = dev ? window.location.hostname + ":7745" : window.location.host;

  // Configure transport endpoints with proper protocols
  // WebSocket uses ws:// or wss://, while SSE and HTTP Streaming use http:// or https://
  const isSecure = window.location.protocol === "https:";
  const wsProtocol = isSecure ? "wss" : "ws";
  const httpProtocol = isSecure ? "https" : "http";

  const transports: TransportEndpoint[] = [
    {
      transport: "websocket",
      endpoint: `${wsProtocol}://${host}/api/v1/ws/events`,
    },
    {
      transport: "sse",
      endpoint: `${httpProtocol}://${host}/api/v1/ws/events`,
    },
  ];

  console.debug("Configured transports:", transports);

  // Get initial token before creating centrifuge instance
  let initialToken = "";
  try {
    console.debug("Fetching initial connection token...");
    initialToken = await getConnectionToken();
    console.debug("Got initial token successfully");
  } catch (err) {
    console.error("Failed to get initial connection token:", err);
    return;
  }

  centrifuge = new Centrifuge(transports, {
    emulationEndpoint: `${httpProtocol}://${host}/api/v1/ws/events`,
    token: initialToken,
    timeout: 100000,
    getToken: async () => {
      try {
        console.debug("Refreshing connection token...");
        return await getConnectionToken();
      } catch (err) {
        console.error("Failed to refresh connection token:", err);
        throw err;
      }
    },

    debug: dev,
  });

  centrifuge.on("connected", ctx => {
    console.debug("connected to server", {
      transport: ctx.transport,
      client: ctx.client,
      data: ctx.data,
    });
  });

  centrifuge.on("disconnected", ctx => {
    console.debug("disconnected from server", ctx);
  });

  centrifuge.on("error", err => {
    console.error("centrifuge error", err);
  });

  // Subscribe to the tenant-specific channel
  if (currentTenantId) {
    const channel = `${currentTenantId}:mutations`;
    const subscription = centrifuge.newSubscription(channel);

    const throttled = new Map<ServerEvent, (m: EventMessage) => void>();
    throttled.set(ServerEvent.LocationMutation, useThrottleFn(onmessage, 1000));
    throttled.set(ServerEvent.ItemMutation, useThrottleFn(onmessage, 1000));
    throttled.set(ServerEvent.TagMutation, useThrottleFn(onmessage, 1000));

    subscription.on("publication", ctx => {
      const pm = ctx.data as EventMessage;
      const fn = throttled.get(pm.event);
      if (fn) {
        fn(pm);
      }
    });

    subscription.on("subscribed", ctx => {
      console.debug("subscribed to channel", channel, ctx);
    });

    subscription.on("error", err => {
      console.error("subscription error", err);
    });

    subscription.subscribe();
  }

  centrifuge.connect();
}

export function onServerEvent(event: ServerEvent, callback: () => void) {
  const prefs = useViewPreferences();
  currentTenantId = prefs.value.collectionId || null;

  if (!watcherSetup) {
    watch(
      () => prefs.value.collectionId,
      newId => {
        currentTenantId = newId || null;
        if (centrifuge) {
          centrifuge.disconnect();
          centrifuge = null;
          // Don't await, let it connect in background
          connect(e => {
            console.debug("received event", e);
            listeners.get(e.event)?.forEach(c => c());
          }).catch(err => {
            console.error("Failed to reconnect after tenant change:", err);
          });
        }
      }
    );
    watcherSetup = true;
  }

  if (centrifuge === null) {
    // Don't await, let it connect in background
    connect(e => {
      console.debug("received event", e);
      listeners.get(e.event)?.forEach(c => c());
    }).catch(err => {
      console.error("Failed to establish connection:", err);
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
