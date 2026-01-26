<script>
  export let type = 'text';
  export let placeholder = '';
  export let value = '';
  export let disabled = false;
  export let required = false;
  export let customClass = '';
  export let icon = null;

  const baseClasses = 'w-full px-4 py-2.5 bg-gray-50 dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-lg text-sm text-gray-900 dark:text-white placeholder-gray-400 dark:placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-purple-500 focus:border-transparent transition-all disabled:opacity-50 disabled:cursor-not-allowed';
</script>

<div class="relative">
  {#if icon}
    <span class="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400">
      {@html icon}
    </span>
  {/if}
  <input
    {type}
    {placeholder}
    {value}
    {disabled}
    {required}
    class={icon ? `${baseClasses} pl-10` : baseClasses}
    {customClass}
    on:input
  />
</div>
