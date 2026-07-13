export function jdbcDriverRemovalConfirmation(driver, profile) {
  if (!driver || !profile) return null;

  return {
    title: '卸载 JDBC 驱动',
    message: `确定卸载 ${driver.name || driver.id} ${profile.version} 吗？`,
    confirmText: '卸载'
  };
}
