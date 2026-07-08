<script>
  export let variant = 'primary'; // primary, secondary, ghost, danger
  export let size = 'md'; // sm, md, lg
  export let disabled = false;
  export let type = 'button';
  export let customClass = '';
  export let icon = null;
  
  const baseClasses = 'inline-flex items-center justify-center rounded-lg font-medium transition-all focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed';
  
  const variantClasses = {
    primary: 'accent-gradient hover:brightness-95 text-white shadow-sm',
    secondary: 'bg-slate-100 hover:bg-slate-200 text-slate-700 focus-visible:ring-slate-500 dark:bg-slate-700 dark:hover:bg-slate-600 dark:text-slate-200',
    ghost: 'bg-transparent hover:bg-slate-100 text-slate-700 focus-visible:ring-slate-500 dark:hover:bg-slate-800 dark:text-slate-200',
    danger: 'bg-rose-600 hover:bg-rose-700 text-white shadow-sm focus-visible:ring-rose-500',
  };
  
  const sizeClasses = {
    sm: 'px-3 py-1.5 text-sm',
    md: 'px-4 py-2.5 text-sm',
    lg: 'px-6 py-3 text-base',
  };

  $: finalClasses = `${baseClasses} ${variantClasses[variant]} ${sizeClasses[size]} ${customClass}`;
</script>

<button
  {type}
  {disabled}
  class={finalClasses}
>
  {#if icon}
    <span class="mr-2">{@html icon}</span>
  {/if}
  <slot />
</button>
