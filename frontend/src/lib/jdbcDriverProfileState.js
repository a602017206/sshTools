export function jdbcProfileActionState({ profileInstalled }) {
  const installed = Boolean(profileInstalled);
  return {
    canInstall: !installed,
    canValidate: installed,
    canRemove: installed
  };
}
