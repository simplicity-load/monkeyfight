{
  pkgs,
  name,
  version,
  ...
}:
with pkgs;
rec {
  default = be;
  be = buildGoModule {
    name = name;
    version = version;
    deleteVendor = true;

    src = ../src/example;
    vendorHash = "sha256-7Wa8tw6sru/SLoPcA8+qxbHvkFr5PQcWZa8wUXm48qY=";
  };
}
