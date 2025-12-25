<script setup lang="ts">
  export type RulerUnit = "metric" | "imperial";

  const props = withDefaults(
    defineProps<{
      width: number; // in mm
      height: number; // in mm
      pixelWidth: number; // canvas width in pixels
      pixelHeight: number; // canvas height in pixels
      unit?: RulerUnit;
    }>(),
    {
      unit: "metric",
    }
  );

  // Size of the ruler in pixels
  const rulerSize = 20;

  // Conversion factor: 1 inch = 25.4mm
  const mmPerInch = 25.4;

  // Unit display label
  const unitLabel = computed(() => (props.unit === "imperial" ? "in" : "mm"));

  // Generate tick marks for horizontal ruler
  const horizontalTicks = computed(() => {
    const ticks: { position: number; label: string; isMajor: boolean }[] = [];
    const mmPerPixel = props.width / props.pixelWidth;

    if (props.unit === "imperial") {
      // Imperial: major ticks every 1", minor every 0.25"
      const widthInches = props.width / mmPerInch;
      for (let inch = 0; inch <= widthInches; inch += 0.25) {
        const mm = inch * mmPerInch;
        const position = mm / mmPerPixel;
        const isMajor = inch === Math.floor(inch); // Whole inches are major
        const isHalf = inch % 0.5 === 0 && !isMajor;
        ticks.push({
          position,
          label: isMajor ? String(Math.round(inch)) : "",
          isMajor: isMajor || isHalf, // Half inches also get longer ticks
        });
      }
    } else {
      // Metric: major ticks every 10mm, minor every 5mm
      for (let mm = 0; mm <= props.width; mm += 5) {
        const position = mm / mmPerPixel;
        const isMajor = mm % 10 === 0;
        ticks.push({
          position,
          label: isMajor ? String(mm) : "",
          isMajor,
        });
      }
    }
    return ticks;
  });

  // Generate tick marks for vertical ruler
  const verticalTicks = computed(() => {
    const ticks: { position: number; label: string; isMajor: boolean }[] = [];
    const mmPerPixel = props.height / props.pixelHeight;

    if (props.unit === "imperial") {
      // Imperial: major ticks every 1", minor every 0.25"
      const heightInches = props.height / mmPerInch;
      for (let inch = 0; inch <= heightInches; inch += 0.25) {
        const mm = inch * mmPerInch;
        const position = mm / mmPerPixel;
        const isMajor = inch === Math.floor(inch);
        const isHalf = inch % 0.5 === 0 && !isMajor;
        ticks.push({
          position,
          label: isMajor ? String(Math.round(inch)) : "",
          isMajor: isMajor || isHalf,
        });
      }
    } else {
      // Metric: major ticks every 10mm, minor every 5mm
      for (let mm = 0; mm <= props.height; mm += 5) {
        const position = mm / mmPerPixel;
        const isMajor = mm % 10 === 0;
        ticks.push({
          position,
          label: isMajor ? String(mm) : "",
          isMajor,
        });
      }
    }
    return ticks;
  });
</script>

<template>
  <div class="relative">
    <!-- Corner square -->
    <div
      class="absolute left-0 top-0 z-10 flex items-center justify-center bg-muted text-[8px] text-muted-foreground"
      :style="{ width: `${rulerSize}px`, height: `${rulerSize}px` }"
    >
      {{ unitLabel }}
    </div>

    <!-- Horizontal ruler -->
    <div
      class="absolute top-0 overflow-hidden bg-muted"
      :style="{ left: `${rulerSize}px`, width: `${pixelWidth}px`, height: `${rulerSize}px` }"
    >
      <svg :width="pixelWidth" :height="rulerSize" class="block">
        <g v-for="tick in horizontalTicks" :key="'h-' + tick.position">
          <line
            :x1="tick.position"
            :y1="tick.isMajor ? 8 : 12"
            :x2="tick.position"
            :y2="rulerSize"
            stroke="currentColor"
            class="text-muted-foreground"
            stroke-width="0.5"
          />
          <text v-if="tick.label" :x="tick.position + 2" y="8" class="fill-muted-foreground text-[8px]" font-size="8">
            {{ tick.label }}
          </text>
        </g>
      </svg>
    </div>

    <!-- Vertical ruler -->
    <div
      class="absolute left-0 overflow-hidden bg-muted"
      :style="{ top: `${rulerSize}px`, width: `${rulerSize}px`, height: `${pixelHeight}px` }"
    >
      <svg :width="rulerSize" :height="pixelHeight" class="block">
        <g v-for="tick in verticalTicks" :key="'v-' + tick.position">
          <line
            :x1="tick.isMajor ? 8 : 12"
            :y1="tick.position"
            :x2="rulerSize"
            :y2="tick.position"
            stroke="currentColor"
            class="text-muted-foreground"
            stroke-width="0.5"
          />
          <text v-if="tick.label" :x="2" :y="tick.position + 8" class="fill-muted-foreground text-[8px]" font-size="8">
            {{ tick.label }}
          </text>
        </g>
      </svg>
    </div>

    <!-- Canvas slot -->
    <div
      :style="{
        marginLeft: `${rulerSize}px`,
        marginTop: `${rulerSize}px`,
      }"
    >
      <slot />
    </div>
  </div>
</template>
