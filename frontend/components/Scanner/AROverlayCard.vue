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
          {{ entity.location.children.length }}
          {{ t("scanner_ar.children", { count: entity.location.children.length }) }}
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
  import { useElementSize, useWindowSize } from "@vueuse/core";
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
  const { width: viewportW } = useWindowSize();

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

    // Apparent QR screen size = average of its top and bottom edges. We floor
    // the card's rendered width so a far-away tiny QR doesn't produce
    // illegible squished text. Close-up QRs are left alone — the card grows
    // with them.
    //
    // 200px is the smallest width where the card's text-sm/text-xs content
    // stays readable after the p-2 padding, sized for the typical user on a
    // Pixel 6 / iPhone 14 in portrait (~390–412 logical px wide). The vw cap
    // keeps it from dominating unusually narrow viewports.
    const apparentSize = (ptDist(tl, tr) + ptDist(bl, br)) / 2;
    const vw = viewportW.value || 1920;
    const minW = Math.min(200, vw * 0.55);
    const targetW = Math.max(minW, apparentSize);
    const scale = apparentSize > 0 ? targetW / apparentSize : 1;

    // Scale all four QR corners around the bottom-edge midpoint. This grows
    // (or shrinks) the QR's projected quad while keeping its anchor and
    // perspective tilt intact, so the card sits where the QR's bottom edge
    // would be if the QR were the clamped target size.
    const anchorX = (bl.x + br.x) / 2;
    const anchorY = (bl.y + br.y) / 2;
    const scalePoint = (p: Point2D): Point2D => ({
      x: anchorX + (p.x - anchorX) * scale,
      y: anchorY + (p.y - anchorY) * scale,
    });
    const sTL = scalePoint(tl);
    const sTR = scalePoint(tr);
    const sBR = scalePoint(br);
    const sBL = scalePoint(bl);

    // Perspective "down" vectors derived from the scaled quad
    const leftDir = { x: sBL.x - sTL.x, y: sBL.y - sTL.y };
    const rightDir = { x: sBR.x - sTR.x, y: sBR.y - sTR.y };
    const leftLen = ptDist(sTL, sBL);
    const rightLen = ptDist(sTR, sBR);

    // Gap between QR bottom edge and card top edge (8 screen px along the
    // perspective down direction)
    const gap = 8;
    const gapL = leftLen > 0 ? gap / leftLen : 0;
    const gapR = rightLen > 0 ? gap / rightLen : 0;

    // Preserve the card's natural aspect ratio at the clamped size — without
    // this, the homography would stretch the card to QR-width × natural-CSS-
    // height and distort the layout.
    const heightRatio = w > 0 ? h / w : 0;

    // Destination corners: card sits below the (virtually-resized) QR on the
    // same perspective plane.
    const dstTL = { x: sBL.x + leftDir.x * gapL, y: sBL.y + leftDir.y * gapL };
    const dstTR = { x: sBR.x + rightDir.x * gapR, y: sBR.y + rightDir.y * gapR };
    const dstBL = {
      x: sBL.x + leftDir.x * (gapL + heightRatio),
      y: sBL.y + leftDir.y * (gapL + heightRatio),
    };
    const dstBR = {
      x: sBR.x + rightDir.x * (gapR + heightRatio),
      y: sBR.y + rightDir.y * (gapR + heightRatio),
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
