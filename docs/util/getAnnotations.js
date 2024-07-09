import { spawn } from "node:child_process";
import { mkdtemp, readFile, existsSync, writeFileSync, mkdirSync } from "node:fs";
import { join } from "node:path";
import { tmpdir } from "node:os";

if (!existsSync(join(process.cwd(), "package.json"))) {
  console.error("error: not running from docs folder");
  process.exit(1);
}

mkdirSync(join(process.cwd(), "assets/data"), { recursive: true });

console.log("running go2json to generate annotations.json");
mkdtemp(join(tmpdir(), "go2json-"), (err, directory) => {
  if (err) throw err;

  const fn = join(directory, "annotations.json");

  const cmd = spawn(
    "go",
    [
      "run",
      "github.com/asty-org/asty@latest",
      "go2json",
      "-comments",
      "-input",
      "../annotations.go",
      "-output",
      fn,
    ],
    {
      windowsHide: true,
      timeout: 90000,
      shell: true,
    }
  );

  cmd.on("close", (code) => {
    if (code !== 0) {
      console.error(`go2json exited with code ${code}`);
      process.exit(code);
    }

    readFile(fn, { encoding: "utf-8" }, (err, input) => {
      if (err) throw err;

      const output = JSON.parse(input);

      let data = {
        annotations: {},
        builders: {},
      };

      let builders = [];

      for (const decl of output.Decls) {
        if (decl.Tok !== "type") continue;

        for (const spec of decl.Specs) {
          if (spec.Name.Name !== "Annotation") continue;

          for (const field of spec.Type.Fields.List) {
            if (!field.Tag.Value) continue;

            const tagMatch = field.Tag.Value.matchAll(/(?<name>[a-z]+):\"(?<value>[^"]+)\"/g);

            let tags = {};
            for (const tag of tagMatch) {
              tags[tag.groups.name] = tag.groups.value.split(",");
            }

            if (tags.builders) builders.push(...tags.builders);

            data.annotations[field.Names[0].Name] = {
              tags: tags,
            };
          }
        }
      }

      for (const builder of builders) {
        for (const decl of output.Decls) {
          if (decl.NodeType != "FuncDecl") continue;
          if (decl.Name.Name != builder) continue;

          data.builders[builder] = {
            comment: decl.Doc.List.map((line) => line.Text.replace(/^\/\/([\s\t]|$)/g, "")),
          };
        }
      }

      const outFn = join(process.cwd(), "assets/data", "annotations.json");
      console.log(`writing ${outFn}`);
      writeFileSync(outFn, JSON.stringify(data, null, 2));
    });
  });
});
