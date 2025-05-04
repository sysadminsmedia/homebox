# Shadcn-Vue

[Shadcn-Vue](https://www.shadcn-vue.com/) is a collection of Vue components based on [shadcn/ui](https://ui.shadcn.com/). We use Shadcn-Vue for our component system.

## What is shadcn-vue?

To quote shadcn-vue:

> This is NOT a component library. It's a collection of re-usable components that you can copy and paste or use the CLI to add to your apps.
> What do you mean not a component library?
> It means you do not install it as a dependency. It is not available or distributed via npm, with no plans to publish it.
> Pick the components you need. Use the CLI to automatically add the components, or copy and paste the code into your project and customize to your needs. The code is yours.

The key advantage of this approach is that we have full control over the components and can modify them to suit our specific needs without being constrained by a third-party dependency.

## Adding Components

1. Add components using the CLI:
   ```bash
   pnpx shadcn-vue@latest add [component-name]
   ```
   For example:
   ```bash
   pnpx shadcn-vue@latest add button
   ```

2. The components will be added to the route in the `components/ui` it then needs to be moved to `frontend/components/ui` for use.

## Usage

1. Import components from the components directory:
   ```vue
   import { Button } from '@/components/ui/button'
   ```

2. Components can be used with their respective props and slots as documented in the shadcn-vue documentation.

## Modifying Components

When modifying components, follow these best practices:

1. If you need to modify a component for a specific use case:
   - Copy the component and give it a name that reflects its purpose
   - Keep the original shadcn component intact for other uses

2. When making global changes:
   - Modify the component in the `components/ui` directory
   - Document any significant changes in comments
