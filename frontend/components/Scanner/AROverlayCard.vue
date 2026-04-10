<template>
  <div
    ref="cardRef"
    class="absolute left-0 top-0 z-10 w-[250px] cursor-pointer rounded-lg border border-border bg-background/85 p-2 shadow-lg backdrop-blur-sm"
    :style="cardStyle"
    @click="handleClick"
  >
    <!-- Loading state -->
    <div v-if="loading" class="flex flex-col gap-1.5">
      <Skeleton class="h-4 w-28" />
      <Skeleton class="h-3 w-20" />
    </div>

    <!-- Error state -->
    <div v-else-if="error || !entity" class="text-xs text-destructive">
      {{ t("scanner_ar.not_found") }}
    </div>

    <!-- Item with optional children -->
    <div v-else-if="entity.item" class="flex flex-col gap-1">
      <span class="truncate text-sm font-semibold text-foreground">{{ entity.item.name }}</span>
      <div class="flex flex-wrap items-center gap-1.5">
        <Badge v-if="entity.item.location" variant="secondary" class="text-xs">
          {{ entity.item.location.name }}
        </Badge>
        <span v-if="entity.item.quantity > 1" class="text-xs text-muted-foreground">
          {{ t("scanner_ar.qty") }}: {{ entity.item.quantity }}
        </span>
      </div>
      <span v-if="entity.item.purchasePrice" class="text-xs text-muted-foreground">
        {{ formattedPrice }}
      </span>
      <!-- Child items -->
      <div v-if="entity.childItems && entity.childItems.length > 0" class="mt-1 border-t border-border pt-1">
        <span class="text-xs font-medium text-muted-foreground">
          {{ entity.childItems.length }} {{ t("scanner_ar.child_items", { count: entity.childItems.length }) }}:
        </span>
        <div
          v-for="child in entity.childItems.slice(0, 3)"
          :key="child.id"
          class="truncate text-xs text-muted-foreground"
        >
          &bull; {{ child.name }}
        </div>
        <span v-if="entity.childItems.length > 3" class="text-xs text-muted-foreground">
          +{{ entity.childItems.length - 3 }} {{ t("scanner_ar.more") }}
        </span>
      </div>
    </div>

    <!-- Location with items -->
    <div v-else-if="entity.location" class="flex flex-col gap-1">
      <span class="truncate text-sm font-semibold text-foreground">{{ entity.location.name }}</span>
      <div class="flex flex-wrap items-center gap-1.5">
        <span v-if="entity.location.children.length > 0" class="text-xs text-muted-foreground">
          {{ entity.location.children.length }} {{ t("scanner_ar.children", { count: entity.location.children.length }) }}
        </span>
      </div>
      <span v-if="entity.location.totalPrice" class="text-xs text-muted-foreground">
        {{ formattedTotalPrice }}
      </span>
      <!-- Items in location -->
      <div v-if="entity.childItems && entity.childItems.length > 0" class="mt-1 border-t border-border pt-1">
        <span class="text-xs font-medium text-muted-foreground">
          {{ entity.childItems.length }} {{ t("scanner_ar.items_in_location", { count: entity.childItems.length }) }}:
        </span>
        <div
          v-for="child in entity.childItems.slice(0, 3)"
          :key="child.id"
          class="truncate text-xs text-muted-foreground"
        >
          &bull; {{ child.name }}
        </div>
        <span v-if="entity.childItems.length > 3" class="text-xs text-muted-foreground">
          +{{ entity.childItems.length - 3 }} {{ t("scanner_ar.more") }}
        </span>
      </div>
    </div>

    <!-- Asset with multiple items (no single parent item) -->
    <div v-else-if="entity.childItems && entity.childItems.length > 0" class="flex flex-col gap-1">
      <span class="text-sm font-semibold text-foreground">
        {{ entity.childItems.length }} {{ t("scanner_ar.items", { count: entity.childItems.length }) }}
      </span>
      <div
        v-for="child in entity.childItems.slice(0, 3)"
        :key="child.id"
        class="truncate text-xs text-muted-foreground"
      >
        &bull; {{ child.name }}
      </div>
      <span v-if="entity.childItems.length > 3" class="text-xs text-muted-foreground">
        +{{ entity.childItems.length - 3 }} {{ t("scanner_ar.more") }}
      </span>
    </div>
  </div>
</template>

<script setup lang="ts">
  import { computed, ref, watchEffect } from "vue";
  import { useElementSize } from "@vueuse/core";
  import { useI18n } from "vue-i18n";
  import type { EntityData, Pose3D, Point2D } from "@/composables/use-barcode-detector";
  import { solveHomography, homographyToMatrix3d } from "@/composables/use-barcode-detector";
  import { Badge } from "@/components/ui/badge";
  import { Skeleton } from "@/components/ui/skeleton";

  const props = defineProps<{
    position: DOMRect;
    pose: Pose3D;
    cornerPoints: Point2D[];
    entity: EntityData | null;
    entityType: "item" | "location" | "asset";
    loading: boolean;
    error: boolean;
  }>();

  const { t } = useI18n();

  const formattedPrice = ref("");
  const formattedTotalPrice = ref("");

  const cardRef = ref<HTMLElement>();
  const { width: cardW, height: cardH } = useElementSize(cardRef);

  function ptDist(a: Point2D, b: Point2D): number {
    return Math.sqrt((a.x - b.x) ** 2 + (a.y - b.y) ** 2);
  }

  const cardStyle = computed(() => {
    const corners = props.cornerPoints;
    const w = cardW.value;
    const h = cardH.value;

    // Need 4 corner points and a measured card size to compute homography
    if (corners.length < 4 || !w || !h) {
      // Fallback: simple translate positioning
      const box = props.position;
      const left = box.x + box.width / 2;
      const top = box.y + box.height + 8;
      return {
        transformOrigin: "0 0",
        transform: `translate(${left}px, ${top}px) translateX(-50%)`,
      };
    }

    const [tl, tr, br, bl] = corners as [Point2D, Point2D, Point2D, Point2D];

    // The QR code's left and right edge vectors define the "down" direction on the QR plane
    const leftDir = { x: bl.x - tl.x, y: bl.y - tl.y };
    const rightDir = { x: br.x - tr.x, y: br.y - tr.y };
    const leftLen = ptDist(tl, bl);
    const rightLen = ptDist(tr, br);

    // Gap between QR bottom edge and card top edge (as fraction of edge length)
    const gap = 8;
    const gapL = leftLen > 0 ? gap / leftLen : 0;
    const gapR = rightLen > 0 ? gap / rightLen : 0;

    // How far to extend for the card's height, proportional to QR size
    const avgEdge = (leftLen + rightLen) / 2 || 1;
    const heightRatio = h / avgEdge;

    // Destination corners: card sits below QR on the same perspective plane
    // Top edge = QR bottom edge + small gap along perspective lines
    // Bottom edge = further along the same perspective lines
    const dstTL = { x: bl.x + leftDir.x * gapL, y: bl.y + leftDir.y * gapL };
    const dstTR = { x: br.x + rightDir.x * gapR, y: br.y + rightDir.y * gapR };
    const dstBL = {
      x: bl.x + leftDir.x * (gapL + heightRatio),
      y: bl.y + leftDir.y * (gapL + heightRatio),
    };
    const dstBR = {
      x: br.x + rightDir.x * (gapR + heightRatio),
      y: br.y + rightDir.y * (gapR + heightRatio),
    };

    // Source: card's natural rect (before transform)
    const src: [Point2D, Point2D, Point2D, Point2D] = [
      { x: 0, y: 0 },
      { x: w, y: 0 },
      { x: w, y: h },
      { x: 0, y: h },
    ];
    const dst: [Point2D, Point2D, Point2D, Point2D] = [dstTL, dstTR, dstBR, dstBL];

    const H = solveHomography(src, dst);
    const matrix = homographyToMatrix3d(H);

    return {
      transformOrigin: "0 0",
      transform: matrix,
    };
  });

  let currencyFormatter: ((value: number | string) => string) | null = null;

  watchEffect(async () => {
    const item = props.entity?.item;
    const location = props.entity?.location;

    if (!item?.purchasePrice && !location?.totalPrice) {
      formattedPrice.value = "";
      formattedTotalPrice.value = "";
      return;
    }

    if (!currencyFormatter) {
      currencyFormatter = await useFormatCurrency();
    }

    formattedPrice.value = item?.purchasePrice ? currencyFormatter(item.purchasePrice) : "";
    formattedTotalPrice.value = location?.totalPrice ? currencyFormatter(location.totalPrice) : "";
  });

  function handleClick() {
    if (props.entity?.item?.id) {
      navigateTo(`/item/${props.entity.item.id}`);
    } else if (props.entity?.location?.id) {
      navigateTo(`/location/${props.entity.location.id}`);
    } else if (props.entity?.childItems && props.entity.childItems.length > 0) {
      navigateTo(`/assets/${props.entity.childItems[0]!.assetId}`);
    }
  }
</script>
