// Fix invalid peerDependency spec in older react-instantsearch-nextjs versions.
// pnpm 10 rejects the `&&` operator in semver ranges; patch it to standard syntax.
function readPackage(pkg) {
  if (pkg.name === 'react-instantsearch-nextjs' && pkg.peerDependencies) {
    if (pkg.peerDependencies.next?.includes('&&')) {
      pkg.peerDependencies.next = '>=13.4.0 <14.0.0';
    }
    if (pkg.peerDependencies['react-instantsearch']?.includes('&&')) {
      pkg.peerDependencies['react-instantsearch'] = '>=7.1.0 <8.0.0';
    }
  }
  return pkg;
}

module.exports = { hooks: { readPackage } };
