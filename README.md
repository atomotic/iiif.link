# iiif.link

A kind of "url shortener" for sharing a view on an IIIF image region (using [Openseadragon](https://openseadragon.github.io/) viewer).

Demo at [https://iiif.link](https://iiif.link): open a manifest, select a page, zoom and pan to desired detail, and click share. A unique URL will be generated that can be shared over the internet; upon opening the URL the viewer will open at the saved zoom and position.

Opengraph meta tags allows a preview of the image region with the manifest label. Try to paste an [example link](https://iiif.link/id/1iDriW37eDJad8SmVPzCW1DYLrJ) into a chat

![](https://docuver.se/tmp/iiif.link-preview.png).

A `HEAD` request on a link display some data with _X-Iiif_ headers

    curl -I https://iiif.link/id/1iDriW37eDJad8SmVPzCW1DYLrJ
    ...
    X-Iiif-Canvas: https://jarvis.edl.beniculturali.it/images/iiif/db/791b6aaf-af7c-4e2e-955a-51715d1c83e0
    X-Iiif-Image: https://jarvis.edl.beniculturali.it/images/iiif/db/791b6aaf-af7c-4e2e-955a-51715d1c83e0/35,168,1703,788/,100/0/default.jpg
    X-Iiif-Label: De Sphaera. Sphaerae coelestis et planetarum descriptio
    X-Iiif-Manifest: https://iiif.edl.beniculturali.it/10965/manifest
    X-Iiif-Page: 11

## QA

Q: Wasn't [IIIF Region](https://iiif.io/api/image/3.0/#41-region) linking enough?  
A: Requesting IIIF images with the region parameter returns only the underlying content as a static image, without the context of the manifest where the image belongs or metadata.

Q: Wasn't [WEB Annotations](https://iiif.io/api/presentation/3.0/#56-annotation) enough?  
A: Annotations cannot be natively shared and referenced with just a browser. You need an external library, like [Annona](https://ncsu-libraries.github.io/annona/) components, to view the annotation.

# Install and run

    go build
    ./iiif.link

Open http://localhost:8080

## Todo

- The viewer is quite simple. Could be improved (also to support Manifest version 3.0)
- Metadata from manifest should be displayed.
- Make `/embed/{id}` to share the view within an iframe
- Make `/id/{id}.json` to export a Web Annotation json
