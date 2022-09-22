Totem is a Multi-user OSINT Tool

Totem currently supports
- VSCO

I plan to add the following services in order:
- Pinterest
- TikTok

Installation:

If you are on macOS, you can download the alpha release. OR you can download the source code and build it yourself if you have the golang toolchain. The Totem directory is on your /Desktop folder by default.
Make a Totem/Tracking.yaml to hold all of the tracking profiles. Here is a default template for a tracking profile:

```yaml
- targetname: BobTheTarget
  active: true
  accounts:
    - userid: -1
      siteid: -1
      username: bob
```

Set userid & siteid to -1 if you don't know them. 

Commands:

`totem run` *Runs all active profiles*

`totem run TargetName...`  *Runs all selected profiles, regardless of whether they are active*

`totem print` *Prints out all profiles*

`totem bio TargetName` *Prints out a profile's bio information*