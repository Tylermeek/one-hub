'use client';

import * as React from 'react';
import * as CollapsiblePrimitive from '@radix-ui/react-collapsible';

const Collapsible = React.forwardRef(({ ...props }, ref) => {
  return <CollapsiblePrimitive.Root ref={ref} data-slot="collapsible" {...props} />;
});

const CollapsibleTrigger = React.forwardRef(({ ...props }, ref) => {
  return <CollapsiblePrimitive.CollapsibleTrigger ref={ref} data-slot="collapsible-trigger" {...props} />;
});

const CollapsibleContent = React.forwardRef(({ ...props }, ref) => {
  return <CollapsiblePrimitive.CollapsibleContent ref={ref} data-slot="collapsible-content" {...props} />;
});

Collapsible.displayName = 'Collapsible';
CollapsibleTrigger.displayName = 'CollapsibleTrigger';
CollapsibleContent.displayName = 'CollapsibleContent';

export { Collapsible, CollapsibleTrigger, CollapsibleContent };
