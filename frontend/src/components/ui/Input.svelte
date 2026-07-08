<script>
  export let type = 'text';
  export let placeholder = '';
  export let value = '';
  export let disabled = false;
  export let required = false;
  export let customClass = '';
  export let icon = null;

  const baseClasses = 'w-full px-4 py-2.5 bg-slate-50 dark:bg-slate-700 border border-slate-200 dark:border-slate-600 rounded-lg text-sm text-slate-900 dark:text-white placeholder-slate-400 dark:placeholder-slate-500 focus:outline-none focus-visible:ring-2 focus:border-transparent transition-all disabled:opacity-50 disabled:cursor-not-allowed';
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
