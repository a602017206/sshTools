<script>
  export let variant = 'primary'; // primary, secondary, ghost, danger
  export let size = 'md'; // sm, md, lg
  export let disabled = false;
  export let type = 'button';
  export let customClass = '';
  export let icon = null;
  
  const baseClasses = 'inline-flex items-center justify-center rounded-lg font-medium transition-all focus:outline-none focus:ring-2 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed';
  
  const variantClasses = {
    primary: 'bg-gradient-to-r from-purple-600 to-blue-600 hover:from-purple-700 hover:to-blue-700 text-white shadow-sm focus:ring-purple-500',
    secondary: 'bg-gray-100 hover:bg-gray-200 text-gray-700 focus:ring-gray-500 dark:bg-gray-700 dark:hover:bg-gray-600 dark:text-gray-200',
    ghost: 'bg-transparent hover:bg-gray-100 text-gray-700 focus:ring-gray-500 dark:hover:bg-gray-800 dark:text-gray-200',
    danger: 'bg-red-600 hover:bg-red-700 text-white shadow-sm focus:ring-red-500',
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
