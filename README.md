Totem is a Multi-user OSINT Tool

Totem currently supports
- VSCO

I plan to add the following services in order:
- Pinterest
- TikTok

Installation:

There is no convenient installation option yet. You can download the project and build if you have the golang toolchain. The Totem directory is on your /Desktop folder by default.
Totem/Tracking.yaml holds all the tracking profiles. Here is a default template for a tracking profile:

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
