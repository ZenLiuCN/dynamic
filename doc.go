/*
Package dynamic is a hot loader toolkit based on [goloader].

# License

Source codes are under Apache License Version 2.0.

# Underwater

 1. Can load and link relocatable object files at runtime, also unload or reload them, in other words, goload is a runtime linker.
 2. Module codes are loaded into an executable Mapping memory section ( same as how other JIT solutions work ).
 3. Having lower footprint as go official plugin.

# Notes

 1. This project is in WIP stage. Current only target on go 1.21+.
 2. User must be careful when use global symbols of those not ship with the host executable, other dynamics may depend on them.
    also there free sequence is important. current user should do it by themselves.
 3. For [goloader]'s limitation, current only exported function can link and use.
 4. Sym must directly fetch and use in code,should not reuse the cast result or the Sym itself. But the function result is safe to use for multiple times.

# Compile tool

This is a module compile tool to compile go files into relocatable object file (extension as .o),
which can be loaded and execute at runtime as a Dynamic module.
The compile tool can be installed by:

	go install github.com/ZenLiuCN/dynamic/compile@latest

It also can inspect the .o file's imports, prepare go sdk to compile the host executable and so on ... .
For more details see the cli help:

	compile -h

# Use this library on develop stage or compile distribution binaries

  - 1. Prepare GO sdk

    use compile cli tool via `compile prepare` or  use shell script named as `patch.sh`.

  - 2. Work around with the dynamics

  - 3. Restore the GO SDK

    use compile cli tool via `compile clean` or use shell script named as `patch.sh`.

# Samples

See testdata and tests.

[goloader]: https://github.com/pkujhd/goloader
*/
package dynamic
