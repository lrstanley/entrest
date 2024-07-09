export function clean(str: string) {
  return str
    .replace(/\t/g, "    ")
    .replace(/github.com\/lrstanley\/entrest\/_examples\/[^/]+/g, "github.com/example/my-project");
}
