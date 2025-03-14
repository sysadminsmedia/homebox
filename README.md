<div align="center">
  <img src="/docs/public/lilbox.svg" height="200"/>
</div>

<h1 align="center" style="margin-top: -10px"> HomeBox </h1>
<p align="center" style="width: 100;">
   <a href="https://homebox.software/en/">Docs</a>
   |
   <a href="https://demo.homebox.software">Demo</a>
   |
   <a href="https://discord.gg/aY4DCkpNA9">Discord</a>
</p>

## What is HomeBox

HomeBox is the inventory and organization system built for the Home User! With a focus on simplicity and ease of use, Homebox is the perfect solution for your home inventory, organization, and management needs. While developing this project, I've tried to keep the following principles in mind:

- _Simple_ - Homebox is designed to be simple and easy to use. No complicated setup or configuration required. Use either a single docker container, or deploy yourself by compiling the binary for your platform of choice.
- _Blazingly Fast_ - Homebox is written in Go, which makes it extremely fast and requires minimal resources to deploy. In general, idle memory usage is less than 50MB for the whole container.
- _Portable_ - Homebox is designed to be portable and run on anywhere. We use SQLite and an embedded Web UI to make it easy to deploy, use, and backup.

# Screenshots
Check out screenshots of the project [here](https://imgur.com/a/5gLWt2j).
You can also try the demo instances of Homebox:
- [Demo](https://demo.homebox.software)
- [Nightly](https://nightly.homebox.software)
- [VNext](https://vnext.homebox.software/)

## Quick Start

[Configuration & Docker Compose](https://homebox.software/en/quick-start.html)

```bash
# If using the rootless image, ensure data 
# folder has correct permissions
mkdir -p /path/to/data/folder
chown 65532:65532 -R /path/to/data/folder
docker run -d \
  --name homebox \
  --restart unless-stopped \
  --publish 3100:7745 \
  --env TZ=Europe/Bucharest \
  --volume /path/to/data/folder/:/data \
  ghcr.io/sysadminsmedia/homebox:latest
# ghcr.io/sysadminsmedia/homebox:latest-rootless
```

<!-- CONTRIBUTING -->

## Contributing

Contributions are what make the open source community such an amazing place to learn, inspire, and create. Any contributions you make are **greatly appreciated**.

If you are not a coder, you can still contribute financially. Financial contributions help me prioritize working on this project over others and helps me know that there is a real demand for project development.

## Help us Translate
We want to make sure that Homebox is available in as many languages as possible. If you are interested in helping us translate Homebox, please help us via our [Weblate instance](https://translate.sysadminsmedia.com/projects/homebox/).

[![Translation status](http://translate.sysadminsmedia.com/widget/homebox/multi-auto.svg)](http://translate.sysadminsmedia.com/engage/homebox/)

## Credits

- Original project by [@hay-kot](https://github.com/hay-kot)
- Logo by [@lakotelman](https://github.com/lakotelman)
