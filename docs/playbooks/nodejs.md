# Node.js Playbooks

Managing `node` versions, packages, caches, etc. is a huge pain in the ass, and so some decisions have been made to help bring some order to it all.

The following conventions are attempted as much as possible:
* Use [`fnm`](https://github.com/Schniz/fnm) to manage node versions.
  * `nvm` is the usual suspect here, but it's a pain to get working with multiple shells and can be slow at times:
    * [`nvm` is not supported via homebrew](https://github.com/nvm-sh/nvm/issues/2914#issuecomment-1279771215)
    * [complicated `fish` shell support](https://github.com/nvm-sh/nvm#fish)
  * After a lot of headaches I did get it sort of working, but to avoid all of this complexity and get better shell support, [Fast Node Manager](https://github.com/Schniz/fnm) is being tried out for its simplicity and support of other formats (like `.nvmrc`).
* Don't mix package managers in a project.
  * Mixing package managers is a recipe for dependency hell; it's best to consolidate on one within a specific project and define all dependencies as best as possible to not rely on global resources. (This isn't unique to Node.js, but is acutely noticeable in the ecosystem.)
* Be conservative in tool/package choices at the global level.
  * Javascript tooling seems to change on a whim, and we want our base environment to be as stable as possible. Being conservative about tool usage at the global level and pushing that complexity into projects/containerized environments should help immensely.
* Try not to install anything globally.
  * With using `fnm`, we should be able to run everything in containers or in isolated environments locally without needing to resort to installing things locally

