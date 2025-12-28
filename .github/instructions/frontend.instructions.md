# Frontend Components & Pages Instructions (`/frontend/`)

## Overview

The frontend is a Nuxt 4 application with Vue 3 and TypeScript. It uses auto-imports for components and composables, file-based routing, and generated TypeScript types from the backend API.

## Directory Structure

```
frontend/
├── components/              # Vue components (auto-imported)
│   ├── Item/               # Item-related components
│   ├── Location/           # Location components
│   ├── Label/              # Label components
│   ├── Form/               # Form components
│   └── ui/                 # Shadcn-vue UI components
├── pages/                  # File-based routes (auto-routing)
│   ├── index.vue           # Home page (/)
│   ├── items.vue           # Items list (/items)
│   ├── item/
│   │   └── [id].vue        # Item detail (/item/:id)
│   ├── locations.vue       # Locations list (/locations)
│   └── profile.vue         # User profile (/profile)
├── composables/            # Vue composables (auto-imported)
│   ├── use-api.ts          # API client wrapper
│   ├── use-auth.ts         # Authentication
│   └── use-user-api.ts     # User API helpers
├── stores/                 # Pinia state management
│   ├── auth.ts             # Auth state
│   └── preferences.ts      # User preferences
├── lib/
│   └── api/
│       └── types/          # Generated TypeScript types (DO NOT EDIT)
├── layouts/                # Layout components
│   └── default.vue         # Default layout
├── locales/                # i18n translations
├── test/                   # Tests (Vitest + Playwright)
└── nuxt.config.ts          # Nuxt configuration
```

## Auto-Imports

### Components

Components in `components/` are **automatically imported** - no import statement needed:

```vue
<!-- components/Item/Card.vue -->
<template>
  <div class="item-card">{{ item.name }}</div>
</template>

<!-- pages/items.vue - NO import needed -->
<template>
  <ItemCard :item="item" />
</template>
```

**Naming convention:** Nested path becomes component name
- `components/Item/Card.vue` → `<ItemCard />`
- `components/Form/TextField.vue` → `<FormTextField />`

### Composables

Composables in `composables/` are **automatically imported**:

```ts
// composables/use-items.ts
export function useItems() {
  const api = useUserApi()
  
  async function getItems() {
    const { data } = await api.items.getAll()
    return data
  }
  
  return { getItems }
}

// pages/items.vue - NO import needed
const { getItems } = useItems()
const items = await getItems()
```

## File-Based Routing

Pages in `pages/` automatically become routes:

```
pages/index.vue              → /
pages/items.vue              → /items
pages/item/[id].vue          → /item/:id
pages/locations.vue          → /locations
pages/location/[id].vue      → /location/:id
pages/profile.vue            → /profile
```

### Dynamic Routes

Use square brackets for dynamic segments:

```vue
<!-- pages/item/[id].vue -->
<script setup lang="ts">
const route = useRoute()
const id = route.params.id

const { data: item } = await useUserApi().items.getOne(id)
</script>

<template>
  <div>
    <h1>{{ item.name }}</h1>
  </div>
</template>
```

## API Integration

### Generated Types

API types are auto-generated from backend Swagger docs:

```ts
// lib/api/types/data-contracts.ts (GENERATED - DO NOT EDIT)
export interface ItemOut {
  id: string
  name: string
  quantity: number
  createdAt: Date | string
  updatedAt: Date | string
}

export interface ItemCreate {
  name: string
  quantity?: number
  locationId?: string
}
```

**Regenerate after backend API changes:**
```bash
task generate  # Runs in backend, updates frontend/lib/api/types/
```

### Using the API Client

The `useUserApi()` composable provides typed API access:

```vue
<script setup lang="ts">
import type { ItemCreate, ItemOut } from '~/lib/api/types/data-contracts'

const api = useUserApi()

// GET all items
const { data: items } = await api.items.getAll({
  q: 'search term',
  page: 1,
  pageSize: 20
})

// GET single item
const { data: item } = await api.items.getOne(itemId)

// POST create item
const newItem: ItemCreate = {
  name: 'New Item',
  quantity: 1
}
const { data: created } = await api.items.create(newItem)

// PUT update item
const { data: updated } = await api.items.update(itemId, {
  quantity: 5
})

// DELETE item
await api.items.delete(itemId)
</script>
```

## Component Patterns

### Standard Vue 3 Composition API

```vue
<script setup lang="ts">
import { ref, computed } from 'vue'
import type { ItemOut } from '~/lib/api/types/data-contracts'

// Props
interface Props {
  item: ItemOut
  editable?: boolean
}
const props = defineProps<Props>()

// Emits
interface Emits {
  (e: 'update', item: ItemOut): void
  (e: 'delete', id: string): void
}
const emit = defineEmits<Emits>()

// State
const isEditing = ref(false)
const localItem = ref({ ...props.item })

// Computed
const displayName = computed(() => {
  return props.item.name.toUpperCase()
})

// Methods
function handleSave() {
  emit('update', localItem.value)
  isEditing.value = false
}
</script>

<template>
  <div class="item-card">
    <h3>{{ displayName }}</h3>
    <p v-if="!isEditing">Quantity: {{ item.quantity }}</p>
    
    <input 
      v-if="isEditing" 
      v-model.number="localItem.quantity"
      type="number"
    />
    
    <button v-if="editable" @click="isEditing = !isEditing">
      {{ isEditing ? 'Cancel' : 'Edit' }}
    </button>
    <button v-if="isEditing" @click="handleSave">Save</button>
  </div>
</template>

<style scoped>
.item-card {
  padding: 1rem;
  border: 1px solid #ccc;
  border-radius: 0.5rem;
}
</style>
```

### Using Pinia Stores

```vue
<script setup lang="ts">
import { useAuthStore } from '~/stores/auth'

const authStore = useAuthStore()

// Access state
const user = computed(() => authStore.user)
const isLoggedIn = computed(() => authStore.isLoggedIn)

// Call actions
async function logout() {
  await authStore.logout()
  navigateTo('/login')
}
</script>
```

### Form Handling

```vue
<script setup lang="ts">
import { useForm } from 'vee-validate'
import type { ItemCreate } from '~/lib/api/types/data-contracts'

const api = useUserApi()

const { values, errors, handleSubmit } = useForm<ItemCreate>({
  initialValues: {
    name: '',
    quantity: 1
  }
})

const onSubmit = handleSubmit(async (values) => {
  try {
    const { data } = await api.items.create(values)
    navigateTo(`/item/${data.id}`)
  } catch (error) {
    console.error('Failed to create item:', error)
  }
})
</script>

<template>
  <form @submit.prevent="onSubmit">
    <input v-model="values.name" type="text" placeholder="Item name" />
    <span v-if="errors.name">{{ errors.name }}</span>
    
    <input v-model.number="values.quantity" type="number" />
    <span v-if="errors.quantity">{{ errors.quantity }}</span>
    
    <button type="submit">Create Item</button>
  </form>
</template>
```

## Styling

### Tailwind CSS

The project uses Tailwind CSS for styling:

```vue
<template>
  <div class="flex items-center justify-between p-4 bg-white rounded-lg shadow-md">
    <h3 class="text-lg font-semibold text-gray-900">{{ item.name }}</h3>
    <span class="text-sm text-gray-500">Qty: {{ item.quantity }}</span>
  </div>
</template>
```

### Shadcn-vue Components

UI components from `components/ui/` (Shadcn-vue):

```vue
<script setup lang="ts">
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader } from '@/components/ui/card'
</script>

<template>
  <Card>
    <CardHeader>
      <h3>{{ item.name }}</h3>
    </CardHeader>
    <CardContent>
      <p>{{ item.description }}</p>
      <Button @click="handleEdit">Edit</Button>
    </CardContent>
  </Card>
</template>
```

## Testing

### Vitest (Unit/Integration)

Tests use Vitest with the backend API running:

```ts
// test/items.test.ts
import { describe, it, expect } from 'vitest'
import { useUserApi } from '~/composables/use-user-api'

describe('Items API', () => {
  it('should create and fetch item', async () => {
    const api = useUserApi()
    
    // Create item
    const { data: created } = await api.items.create({
      name: 'Test Item',
      quantity: 1
    })
    
    expect(created.name).toBe('Test Item')
    
    // Fetch item
    const { data: fetched } = await api.items.getOne(created.id)
    expect(fetched.id).toBe(created.id)
  })
})
```

**Run tests:**
```bash
task ui:watch  # Watch mode
cd frontend && pnpm run test:ci  # CI mode
```

### Playwright (E2E)

E2E tests in `test/`:

```ts
// test/e2e/items.spec.ts
import { test, expect } from '@playwright/test'

test('should create new item', async ({ page }) => {
  await page.goto('/items')
  
  await page.click('button:has-text("New Item")')
  await page.fill('input[name="name"]', 'Test Item')
  await page.fill('input[name="quantity"]', '5')
  await page.click('button:has-text("Save")')
  
  await expect(page.locator('text=Test Item')).toBeVisible()
})
```

**Run E2E tests:**
```bash
task test:e2e  # Full E2E suite
```

## Adding a New Feature

### 1. Update Backend API

Make backend changes first (schema, service, handler):
```bash
# Edit backend files
task generate  # Regenerates TypeScript types
```

### 2. Create Component

Create `components/MyFeature/Card.vue`:
```vue
<script setup lang="ts">
import type { MyFeatureOut } from '~/lib/api/types/data-contracts'

interface Props {
  feature: MyFeatureOut
}
defineProps<Props>()
</script>

<template>
  <div>{{ feature.name }}</div>
</template>
```

### 3. Create Page

Create `pages/my-feature/[id].vue`:
```vue
<script setup lang="ts">
const route = useRoute()
const api = useUserApi()

const { data: feature } = await api.myFeature.getOne(route.params.id)
</script>

<template>
  <MyFeatureCard :feature="feature" />
</template>
```

### 4. Test

```bash
task ui:check    # Type checking
task ui:fix      # Linting
task ui:watch    # Run tests
```

## Critical Rules

1. **Never edit generated types** - `lib/api/types/` is auto-generated, run `task generate` after backend changes
2. **No manual imports for components/composables** - auto-imported from `components/` and `composables/`
3. **Use TypeScript** - all `.vue` files use `<script setup lang="ts">`
4. **Follow file-based routing** - pages in `pages/` become routes automatically
5. **Use `useUserApi()` for API calls** - provides typed, authenticated API client
6. **Max 1 linting warning in CI** - run `task ui:fix` before committing
7. **Test with backend running** - integration tests need API server

## Common Issues

- **"Type not found"** → Run `task generate` to regenerate types from backend
- **Component not found** → Check naming (nested path = component name)
- **API call fails** → Ensure backend is running (`task go:run`)
- **Lint errors** → Run `task ui:fix` to auto-fix
- **Type errors** → Run `task ui:check` for detailed errors
