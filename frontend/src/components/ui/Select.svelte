<script>
  export let options = [];
  export let value = '';
  export let placeholder = '请选择...';
  export let disabled = false;
  export let customClass = '';

  const baseClasses = 'w-full px-4 py-2.5 bg-gray-50 dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-lg text-sm text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-purple-500 focus:border-transparent transition-all disabled:opacity-50 disabled:cursor-not-allowed appearance-none bg-no-repeat bg-right';
</script>

<div class="relative">
  <select
    {value}
    {disabled}
    class={baseClasses + ' ' + customClass}
    style="background-image: url('data:image/svg+xml;charset=UTF-8,%3csvg xmlns=\\'http://www.w3.org/2000/svg\\' viewBox=\\'0 0 24 24\\' fill=\\'none\\' stroke=\\'%236b7280\\' stroke-width=\\'2\\' stroke-linecap=\\'round\\' stroke-linejoin=\\'round\\'%3e%3cpolyline points=\\'6 9 12 15 9\\'%3e%3c/polyline%3e%3c/svg%3e'); background-position: right 0.75rem center; background-size: 1rem;"
  >
    {#if placeholder}
      <option value="">{placeholder}</option>
    {/if}
    {#each options as option (option.value)}
      <option value={option.value}>{option.label}</option>
    {/each}
  </select>
</div>
