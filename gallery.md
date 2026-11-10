# GenerativeArt Gallery & Code Examples

[← Back to README.md](./README.md) · [🌐 Live Interactive Gallery (GitHub Pages)](https://jdxyw.github.io/generativeart/) · [📚 API & Parameter Docs](./docs/doc.md)

This gallery contains standalone `Go` usage examples and rendered outputs for `generativeart` (newest waves listed first). You can also browse, filter, search, and copy snippets interactively at **[https://jdxyw.github.io/generativeart/](https://jdxyw.github.io/generativeart/)**.

---

## 🎞️ Seamless 600×600 HD Looping GIF Animations (`DrawTimelapseFrames` + `ToGIF`)

| Brazilian Agate Geode (`agategeodecrystal`) | Fabergé Guilloché Star (`fabergeguillocheornament`) | Polar Aurora Fold (`auroracurtainfold`) | Fidenza Velvet Loom (`fidenzavelvetloom`) |
| :---: | :---: | :---: | :---: |
| ![Agate Geode Crystal Loop](images/agategeodecrystal_loop.gif) | ![Faberge Guilloche Ornament Loop](images/fabergeguillocheornament_loop.gif) | ![Aurora Curtain Fold Loop](images/auroracurtainfold_loop.gif) | ![Fidenza Velvet Loom Loop](images/fidenzavelvetloom_loop.gif) |
| **Abyssal Medusa (`kuroshiobioluminescence`)** | **3D Lorentz Ribbons (`lorentzstrangeribbon`)** | **Solar Eclipse Corona (`coronalmagneticarcade`)** | **3D Hopf Torus Silk (`hopftorussilk`)** |
| ![Kuroshio Bioluminescence Loop](images/kuroshiobioluminescence_loop.gif) | ![Lorentz Strange Ribbon Loop](images/lorentzstrangeribbon_loop.gif) | ![Coronal Magnetic Arcade Loop](images/coronalmagneticarcade_loop.gif) | ![Hopf Torus Silk Loop](images/hopftorussilk_loop.gif) |
| **Volcanic Atoll Terraces (`volcanicatollterraces`)** | **Peacock Ocelli Plumes (`peacockocelliplumes`)** | **Cloisonné Star-Rose (`byzantinerosemosaic`)** | **Murano Glass Nautilus (`muranoglassnautilus`)** |
| ![Volcanic Atoll Terraces Loop](images/volcanicatollterraces_loop.gif) | ![Peacock Ocelli Plumes Loop](images/peacockocelliplumes_loop.gif) | ![Byzantine Rose Mosaic Loop](images/byzantinerosemosaic_loop.gif) | ![Murano Glass Nautilus Loop](images/muranoglassnautilus_loop.gif) |

```bash
go run ./cmd/genart render -art agategeodecrystal        -w 600 -h 600 -frames 20 -delay 12 -out images/agategeodecrystal_loop.gif
go run ./cmd/genart render -art fabergeguillocheornament -w 600 -h 600 -frames 20 -delay 12 -out images/fabergeguillocheornament_loop.gif
go run ./cmd/genart render -art auroracurtainfold        -w 600 -h 600 -frames 20 -delay 12 -out images/auroracurtainfold_loop.gif
go run ./cmd/genart render -art fidenzavelvetloom       -w 600 -h 600 -frames 20 -delay 12 -out images/fidenzavelvetloom_loop.gif
go run ./cmd/genart render -art kuroshiobioluminescence -w 600 -h 600 -frames 20 -delay 12 -out images/kuroshiobioluminescence_loop.gif
go run ./cmd/genart render -art lorentzstrangeribbon    -w 600 -h 600 -frames 20 -delay 12 -out images/lorentzstrangeribbon_loop.gif
go run ./cmd/genart render -art coronalmagneticarcade   -w 600 -h 600 -frames 20 -delay 12 -out images/coronalmagneticarcade_loop.gif
go run ./cmd/genart render -art hopftorussilk         -w 600 -h 600 -frames 20 -delay 12 -out images/hopftorussilk_loop.gif
go run ./cmd/genart render -art volcanicatollterraces -w 600 -h 600 -frames 20 -delay 12 -out images/volcanicatollterraces_loop.gif
go run ./cmd/genart render -art peacockocelliplumes   -w 600 -h 600 -frames 20 -delay 12 -out images/peacockocelliplumes_loop.gif
go run ./cmd/genart render -art byzantinerosemosaic   -w 600 -h 600 -frames 20 -delay 12 -out images/byzantinerosemosaic_loop.gif
go run ./cmd/genart render -art muranoglassnautilus   -w 600 -h 600 -frames 20 -delay 12 -out images/muranoglassnautilus_loop.gif
go run ./cmd/genart render -art haeckelradiolaria     -w 600 -h 600 -frames 20 -delay 12 -out images/haeckelradiolaria_loop.gif
go run ./cmd/genart render -art relativisticlens      -w 600 -h 600 -frames 20 -delay 12 -out images/relativisticlens_loop.gif
go run ./cmd/genart render -art weingartentorusknot   -w 600 -h 600 -frames 20 -delay 12 -out images/weingartentorusknot_loop.gif
go run ./cmd/genart render -art babellibrary          -w 600 -h 600 -frames 20 -delay 12 -out images/babellibrary_loop.gif
go run ./cmd/genart render -art keplerharmonices      -w 600 -h 600 -frames 20 -delay 12 -out images/keplerharmonices_loop.gif
go run ./cmd/genart render -art moebiushermeticgarage -w 600 -h 600 -frames 20 -delay 12 -out images/moebiushermeticgarage_loop.gif
go run ./cmd/genart render -art calabiyaumanifold     -w 600 -h 600 -frames 20 -delay 12 -out images/calabiyaumanifold_loop.gif
go run ./cmd/genart render -art durerpolyhedron       -w 600 -h 600 -frames 20 -delay 12 -out images/durerpolyhedron_loop.gif
go run ./cmd/genart render -art piranesicarceri       -w 600 -h 600 -frames 20 -delay 12 -out images/piranesicarceri_loop.gif
go run ./cmd/genart render -art moebiusleviathan      -w 600 -h 600 -frames 20 -delay 12 -out images/moebiusleviathan_loop.gif
go run ./cmd/genart render -art moebiuscanyon         -w 600 -h 600 -frames 20 -delay 12 -out images/moebiuscanyon_loop.gif
go run ./cmd/genart render -art moebiusincal          -w 600 -h 600 -frames 20 -delay 12 -out images/moebiusincal_loop.gif
go run ./cmd/genart render -art eschercirclelimit     -w 600 -h 600 -frames 20 -delay 12 -out images/eschercirclelimit_loop.gif
go run ./cmd/genart render -art escherdrostegallery   -w 600 -h 600 -seed 20262002 -frames 20 -delay 12 -out images/escherdrostegallery_loop.gif
go run ./cmd/genart render -art escherdrostegallery   -w 600 -h 600 -seed 20262011 -frames 20 -delay 12 -out images/escherdrostewave_loop.gif
go run ./cmd/genart render -art gaborwaveinterference -w 600 -h 600 -frames 20 -delay 12 -out images/gaborwaveinterference_loop.gif
go run ./cmd/genart render -art analyticalcurlsilk    -w 600 -h 600 -frames 20 -delay 12 -out images/analyticalcurlsilk_loop.gif
go run ./cmd/genart render -art nebulaplumes          -w 600 -h 600 -frames 20 -delay 12 -out images/nebulaplumes_loop.gif
go run ./cmd/genart render -art abyssalcurrents     -w 600 -h 600 -frames 20 -delay 12 -out images/abyssalcurrents_loop.gif
go run ./cmd/genart render -art solarprominence     -w 600 -h 600 -frames 20 -delay 12 -out images/solarprominence_loop.gif
go run ./cmd/genart render -art botanicalwhorls     -w 600 -h 600 -frames 20 -delay 12 -out images/botanicalwhorls_loop.gif
go run ./cmd/genart render -art vorteximpasto       -w 600 -h 600 -frames 20 -delay 12 -out images/vorteximpasto_loop.gif
go run ./cmd/genart render -art squeegeescrape      -w 600 -h 600 -frames 20 -delay 12 -out images/squeegeescrape_loop.gif
go run ./cmd/genart render -art basaltcolumns       -w 600 -h 600 -frames 20 -delay 12 -out images/basaltcolumns_loop.gif
go run ./cmd/genart render -art wireframeterrain    -w 600 -h 600 -frames 20 -delay 12 -out images/wireframeterrain_loop.gif
go run ./cmd/genart render -art lissajousribbon     -w 600 -h 600 -frames 20 -delay 12 -out images/lissajousribbon_loop.gif
go run ./cmd/genart render -art marchingisolines    -w 600 -h 600 -frames 20 -delay 12 -out images/marchingisolines_loop.gif
go run ./cmd/genart render -art vortexfluid         -w 600 -h 600 -frames 20 -delay 12 -out images/vortexfluid_loop.gif
go run ./cmd/genart render -art dunhuangapsara      -w 600 -h 600 -frames 16 -delay 7 -out images/dunhuangapsara_loop.gif
go run ./cmd/genart render -art stainedrosewindow   -w 600 -h 600 -frames 16 -delay 7 -out images/stainedrosewindow_loop.gif
go run ./cmd/genart render -art newtonbasin         -w 600 -h 600 -frames 16 -delay 7 -out images/newtonbasin_loop.gif
go run ./cmd/genart render -art origamitessellation -w 600 -h 600 -frames 16 -delay 7 -out images/origamitessellation_loop.gif
go run ./cmd/genart render -art cyclicautomata      -w 600 -h 600 -frames 16 -delay 7 -out images/cyclicautomata_loop.gif
go run ./cmd/genart render -art gyroidsurface       -w 600 -h 600 -frames 16 -delay 7 -out images/gyroidsurface_loop.gif
go run ./cmd/genart render -art mosaicsun           -w 600 -h 600 -frames 16 -delay 7 -out images/mosaicsun_loop.gif
go run ./cmd/genart render -art magneticdipole      -w 600 -h 600 -frames 16 -delay 7 -out images/magneticdipole_loop.gif
```

---

## 🖼️ Museum Exhibition Posters & Multi-Panel Contact Sheets (`genart poster`)

| Curated Multi-Artwork Exhibition (`theme=dark`) | Multi-Seed Variation Study (`theme=light`) |
| :---: | :---: |
| ![Wave 12 Dark Poster](images/poster_wave12_dark.png) | ![Voronoi Light Poster](images/poster_voronoi_light.png) |

```bash
# Curated 2x2 Multi-Engine Exhibition Poster (Obsidian Dark & Gold Mat)
go run ./cmd/genart poster -arts dunhuangapsara,celticknotwork,gyroidsurface,persiantilework -grid 2x2 -panel 360 -theme dark -title "HERITAGE & TPMS SCULPTURE" -out images/poster_wave12_dark.png

# 2x2 Multi-Seed Variation Contact Sheet (Museum Ivory Parchment Mat)
go run ./cmd/genart poster -art voronoicrystal -grid 2x2 -panel 360 -theme light -title "VORONOI AGATE VARIATIONS" -out images/poster_voronoi_light.png
```

---

## Table of Contents

- [Wave 27: Multi-Paradigm Non-Mœbius Flagship Quartet — Brazilian Agate & Amethyst Geode, Song Dynasty Blue-and-Green Shanshui Scroll, Imperial Fabergé Guilloché Enamel Star & 3D Polar Aurora Borealis Fjord](#wave-27-multi-paradigm-non-mœbius-flagship-quartet--brazilian-agate--amethyst-geode-song-dynasty-blue-and-green-shanshui-scroll-imperial-fabergé-guilloché-enamel-star--3d-polar-aurora-borealis-fjord)
  - [Agate Geode Crystal](#agate-geode-crystal)
  - [Song Blue Green Scroll](#song-blue-green-scroll)
  - [Faberge Guilloche Ornament](#faberge-guilloche-ornament)
  - [Aurora Curtain Fold](#aurora-curtain-fold)
- [Wave 26: Multi-Paradigm Non-Mœbius Flagship Quartet — Fidenza Velvet & Gold Loom, Kuroshio Bioluminescence LIC, 3D Lorentz Strange-Attractor Ribbons & Total Solar Eclipse Coronal Arcade](#wave-26-multi-paradigm-non-mœbius-flagship-quartet--fidenza-velvet--gold-loom-kuroshio-bioluminescence-lic-3d-lorentz-strange-attractor-ribbons--total-solar-eclipse-coronal-arcade)
  - [Fidenza Velvet Loom](#fidenza-velvet-loom)
  - [Kuroshio Bioluminescence](#kuroshio-bioluminescence)
  - [Lorentz Strange Ribbon](#lorentz-strange-ribbon)
  - [Coronal Magnetic Arcade](#coronal-magnetic-arcade)
- [Wave 25: Multi-Paradigm Non-Mœbius Flagship Quartet — 3D Hopf Fibration Silk Ribbons, 3D Laser-Cut Volcanic Atoll Terraces, 3D G-Buffer Peacock Ocelli Plumes & Cloisonné Star-Rose Mosaic](#wave-25-multi-paradigm-non-mœbius-flagship-quartet--3d-hopf-fibration-silk-ribbons-3d-laser-cut-volcanic-atoll-terraces-3d-g-buffer-peacock-ocelli-plumes--cloisonné-star-rose-mosaic)
  - [Hopf Torus Silk](#hopf-torus-silk)
  - [Volcanic Atoll Terraces](#volcanic-atoll-terraces)
  - [Peacock Ocelli Plumes](#peacock-ocelli-plumes)
  - [Byzantine Rose Mosaic](#byzantine-rose-mosaic)
- [Wave 24: Phase 2–5 Flagship Quartet — Venetian Murano Blown-Glass Nautilus, Haeckel 3D Silica Radiolaria, Relativistic Black-Hole Gravitational Lens & Weingarten Torus-Knot Sculpture](#wave-24-phase-25-flagship-quartet--venetian-murano-blown-glass-nautilus-haeckel-3d-silica-radiolaria-relativistic-black-hole-gravitational-lens--weingarten-torus-knot-sculpture)
  - [Murano Glass Nautilus](#murano-glass-nautilus)
  - [Haeckel Radiolaria](#haeckel-radiolaria)
  - [Relativistic Lens](#relativistic-lens)
  - [Weingarten Torus Knot](#weingarten-torus-knot)
- [Wave 23: Master Architecture, Celestial Mechanics & Sci-Fi Megastructures — Borges Library of Babel, Kepler Mysterium Cosmographicum, Mœbius Hermetic Garage & Calabi-Yau Algebraic Manifold](#wave-23-master-architecture-celestial-mechanics--sci-fi-megastructures--borges-library-of-babel-kepler-mysterium-cosmographicum-mœbius-hermetic-garage--calabi-yau-algebraic-manifold)
  - [Babel Library](#babel-library)
  - [Kepler Harmonices](#kepler-harmonices)
  - [Moebius Hermetic Garage](#moebius-hermetic-garage)
  - [Calabi Yau Manifold](#calabi-yau-manifold)
- [Wave 22: Part VIII Flagship Quartet — Piranesi Carceri Vault Labyrinth, 3D SDF Mœbius Leviathan Airship, Dürer/Jamnitzer 3D Polyhedron & Kubelka-Munk Mineral Shan-Shui](#wave-22-part-viii-flagship-quartet--piranesi-carceri-vault-labyrinth-3d-sdf-mœbius-leviathan-airship-dürerjamnitzer-3d-polyhedron--kubelka-munk-mineral-shan-shui)
  - [Piranesi Carceri](#piranesi-carceri)
  - [Moebius Leviathan](#moebius-leviathan)
  - [Durer Polyhedron](#durer-polyhedron)
  - [Mineral Shan Shui](#mineral-shan-shui)
- [Wave 21: Jean "Mœbius" Giraud Ligne Claire Quartet — Stratified Desert Canyon, L'Incal Megastructure Sanctuary, Gardens of Edena Travertine Oasis & Starwatcher Dyson-Torus Citadel](#wave-21-jean-mœbius-giraud-ligne-claire-quartet--stratified-desert-canyon-lincal-megastructure-sanctuary-gardens-of-edena-travertine-oasis--starwatcher-dyson-torus-citadel)
  - [Moebius Canyon](#moebius-canyon)
  - [Moebius Incal](#moebius-incal)
  - [Moebius Edena](#moebius-edena)
  - [Moebius Starwatcher](#moebius-starwatcher)
- [Wave 20: M.C. Escher Trio — Circle Limit Hyperbolic Tessellation, Print Gallery Droste Spiral & Isohedral Metamorphosis](#wave-20-mc-escher-trio--circle-limit-hyperbolic-tessellation-print-gallery-droste-spiral--isohedral-metamorphosis)
  - [Escher Circle Limit](#escher-circle-limit)
  - [Escher Droste Gallery](#escher-droste-gallery)
  - [Escher Metamorphosis](#escher-metamorphosis)
- [Wave 19: Analytical Perlin & Gabor Noise Quartet — Fidenza Flow Blocks, 3D Laser-Cut Topographic Terraces, 3D Perlin Silk Ribbon Sculpture & Celestial Gabor-Curl Astrolabe](#wave-19-analytical-perlin--gabor-noise-quartet--fidenza-flow-blocks-3d-laser-cut-topographic-terraces-3d-perlin-silk-ribbon-sculpture--celestial-gabor-curl-astrolabe)
  - [Perlin Malachite (Fidenza Flow Blocks)](#perlin-malachite)
  - [Erosion Canyon (3D Topographic Terraces)](#erosion-canyon)
  - [Analytical Curl Silk (3D Ribbon Sculpture)](#analytical-curl-silk)
  - [Gabor Wave Interference (Celestial Astrolabe)](#gabor-wave-interference)
- [Wave 18: LIC Jovian Magnetosphere, Thin-Film Iridescent Nacre, Kissing-Schottky Kleinian Pearls & 3D Bismuth Hopper Crystals](#wave-18-lic-jovian-magnetosphere-thin-film-iridescent-nacre-kissing-schottky-kleinian-pearls--3d-bismuth-hopper-crystals)
  - [Jupiter Magnetosphere](#jupiter-magnetosphere)
  - [Iridescent Nacre](#iridescent-nacre)
  - [Schottky Kleinian](#schottky-kleinian)
  - [Bismuth Hopper](#bismuth-hopper)
- [Wave 17: Tactile Bristle & Vortex Impasto Series — Binary Nebula Plumes, Abyssal Kuroshio Currents, Solar Prominence Tondo & Botanical Lotus Whorls](#wave-17-tactile-bristle--vortex-impasto-series--binary-nebula-plumes-abyssal-kuroshio-currents-solar-prominence-tondo--botanical-lotus-whorls)
  - [Nebula Plumes](#nebula-plumes)
  - [Abyssal Currents](#abyssal-currents)
  - [Solar Prominence](#solar-prominence)
  - [Botanical Whorls](#botanical-whorls)
- [Wave 16: Molnár Concentric Morph, Swirling Vortex Impasto, Mid-Century Sliced Pebbles, Hexagonal Mandala Honeycomb, Richter Squeegee Scrape, Retro Sci-Fi Halftone Panels & 3D Fluted Nautilus Spiral](#wave-16-molnár-concentric-morph-swirling-vortex-impasto-mid-century-sliced-pebbles-hexagonal-mandala-honeycomb-richter-squeegee-scrape-retro-sci-fi-halftone-panels--3d-fluted-nautilus-spiral)
  - [Concentric Morph](#concentric-morph)
  - [Vortex Impasto](#vortex-impasto)
  - [Sliced Pebbles](#sliced-pebbles)
  - [Hex Mandala](#hex-mandala)
  - [Squeegee Scrape](#squeegee-scrape)
  - [Halftone Panels](#halftone-panels)
  - [Nautilus Spiral (3D Fluted Ribbon Edition)](#nautilus-spiral-3d-fluted-ribbon-edition)
- [Wave 15: Hydraulic Erosion Alpine Peaks, Isometric Basalt Columns, Aeolian Sand Dunes, Bathymetric Trench Chart & 3D Wireframe Glacier Mesh](#wave-15-hydraulic-erosion-alpine-peaks-isometric-basalt-columns-aeolian-sand-dunes-bathymetric-trench-chart--3d-wireframe-glacier-mesh)
  - [Hydraulic Erosion](#hydraulic-erosion)
  - [Basalt Columns](#basalt-columns)
  - [Sand Dunes](#sand-dunes)
  - [Bathymetric Chart](#bathymetric-chart)
  - [Wireframe Terrain](#wireframe-terrain)
- [Wave 14: Riemann Zeta Critical Strip, 3D Hopf Fibration, Thomas Labyrinth Attractor, 12-Fold Faraday Quasicrystal Waves, Conformal Möbius Loxodrome & FTLE Lagrangian Coherent Structures](#wave-14-riemann-zeta-critical-strip-3d-hopf-fibration-thomas-labyrinth-attractor-12-fold-faraday-quasicrystal-waves-conformal-möbius-loxodrome--ftle-lagrangian-coherent-structures)
  - [Riemann Zeta](#riemann-zeta)
  - [Hopf Fibration](#hopf-fibration)
  - [Thomas Attractor](#thomas-attractor)
  - [Faraday Waves](#faraday-waves)
  - [Mobius Loxodrome](#mobius-loxodrome)
  - [Lagrangian Coherent](#lagrangian-coherent)
- [Wave 13: 3D Halvorsen Attractor, Kleinian Fuchsian Pearls, Abelian Sandpile, Kuramoto Phase Defects, Farey Ford Circles & Quantum Chaos Scars (100-Art Milestone)](#wave-13-3d-halvorsen-attractor-kleinian-fuchsian-pearls-abelian-sandpile-kuramoto-phase-defects-farey-ford-circles--quantum-chaos-scars-100-art-milestone)
  - [Halvorsen Attractor](#halvorsen-attractor)
  - [Hyperbolic Fuchsian](#hyperbolic-fuchsian)
  - [Abelian Sandpile](#abelian-sandpile)
  - [Kuramoto Wave](#kuramoto-wave)
  - [Ford Circles](#ford-circles)
  - [Quantum Billiard](#quantum-billiard)
- [Wave 12: Dunhuang Caisson, Celtic Knotwork, 3D Gyroid TPMS, Lichtenberg Plasma, Persian Muqarnas & Warring States Bronze](#wave-12-dunhuang-caisson-celtic-knotwork-3d-gyroid-tpms-lichtenberg-plasma-persian-muqarnas--warring-states-bronze)
  - [Dunhuang Apsara](#dunhuang-apsara)
  - [Celtic Knotwork](#celtic-knotwork)
  - [Gyroid Surface](#gyroid-surface)
  - [Lichtenberg Tree](#lichtenberg-tree)
  - [Persian Tilework](#persian-tilework)
  - [Champleve Enamel](#champleve-enamel)
- [Wave 11: Jovian Vortex Fluid, BZ Spiral Waves, Gilded PCB WFC, Cursive Calligraphy, Buddhabrot & Lyapunov](#wave-11-jovian-vortex-fluid-bz-spiral-waves-gilded-pcb-wfc-cursive-calligraphy-buddhabrot--lyapunov)
  - [Vortex Fluid](#vortex-fluid)
  - [Cyclic Automata](#cyclic-automata)
  - [Circuit WFC](#circuit-wfc)
  - [Asemic Calligraphy](#asemic-calligraphy)
  - [Buddhabrot](#buddhabrot)
  - [Lyapunov Fractal](#lyapunov-fractal)
- [Wave 10: 8-Fold Quasicrystals, Sashiko Embroidery, Electromagnetic Dipoles & Complex Newton Fractals](#wave-10-8-fold-quasicrystals-sashiko-embroidery-electromagnetic-dipoles--complex-newton-fractals)
  - [Ammann Beenker](#ammann-beenker)
  - [Sashiko Weave](#sashiko-weave)
  - [Magnetic Dipole](#magnetic-dipole)
  - [Newton Basin](#newton-basin)
- [Wave 9: Topological Contours, Streamline Engraving, 3D Origami & Bauhaus SDF Relief](#wave-9-topological-contours-streamline-engraving-3d-origami--bauhaus-sdf-relief)
  - [Even Streamlines](#even-streamlines)
  - [Marching Isolines](#marching-isolines)
  - [Origami Tessellation](#origami-tessellation)
  - [Mondrian Subdivision](#mondrian-subdivision)
- [Wave 8: Decorative Arts, Cloisonné Enamel & Printmaking](#wave-8-decorative-arts-cloisonné-enamel--printmaking)
  - [Cloisonne Enamel](#cloisonne-enamel)
  - [Woodblock Wave](#woodblock-wave)
  - [Art Deco Sunburst](#art-deco-sunburst)
  - [Stained Rose Window](#stained-rose-window)
- [Wave 7: Anisotropic Flow Mosaic & Low-Poly Crystal Series](#wave-7-anisotropic-flow-mosaic--low-poly-crystal-series)
  - [Mosaic Vortex](#mosaic-vortex)
  - [Mosaic Wave](#mosaic-wave)
  - [Mosaic Canyon](#mosaic-canyon)
  - [Mosaic Lotus](#mosaic-lotus)
  - [Mosaic Sun](#mosaic-sun)
- [Wave 6: G-Buffer Impasto, Voxel Terrain & Geometric Replications](#wave-6-g-buffer-impasto-voxel-terrain--geometric-replications)
  - [Impasto Flow](#impasto-flow)
  - [Voxel Terrain](#voxel-terrain)
  - [Packed Capsules](#packed-capsules)
  - [Orphic Discs](#orphic-discs)
  - [Nautilus Spiral](#nautilus-spiral)
  - [Tangent Cones](#tangent-cones)
- [Wave 5: Hyperbolic Geometry, Quasicrystals & 3D DoF Chaos](#wave-5-hyperbolic-geometry-quasicrystals--3d-dof-chaos)
  - [Poincare Disk](#poincare-disk)
  - [Suminagashi](#suminagashi)
  - [Aizawa Attractor](#aizawa-attractor)
  - [Penrose Tiling](#penrose-tiling)
  - [Isometric City](#isometric-city)
  - [DLA Frost](#dla-frost)
- [Wave 4: Fractal Flames, Flocking & Sacred Geometry](#wave-4-fractal-flames-flocking--sacred-geometry)
  - [Fractal Flame](#fractal-flame)
  - [Joy Division](#joy-division)
  - [Boids Flock](#boids-flock)
  - [Islamic Star](#islamic-star)
  - [Delaunay Stained](#delaunay-stained)
  - [Phyllotaxis](#phyllotaxis)
- [Wave 3: Reaction-Diffusion, Harmonographs & Venation](#wave-3-reaction-diffusion-harmonographs--venation)
  - [Gray Scott](#gray-scott)
  - [Lissajous Ribbon](#lissajous-ribbon)
  - [Apollonian Gasket](#apollonian-gasket)
  - [Space Colonization](#space-colonization)
  - [Truchet Flow](#truchet-flow)
  - [Aurora Wave](#aurora-wave)
- [Wave 2: Strange Attractors, Slime Molds & Shan-Shui Ink](#wave-2-strange-attractors-slime-molds--shan-shui-ink)
  - [Clifford Nebula](#clifford-nebula)
  - [Physarum](#physarum)
  - [Differential Growth](#differential-growth)
  - [Chladni](#chladni)
  - [Voronoi Crystal](#voronoi-crystal)
  - [Ink Landscape](#ink-landscape)
- [Wave 1: Curl Flow, Sand Painting, Watercolor & Guilloche](#wave-1-curl-flow-sand-painting-watercolor--guilloche)
  - [Curl Flow](#curl-flow)
  - [Sand Ring](#sand-ring)
  - [Watercolor](#watercolor)
  - [Guilloche](#guilloche)
- [Classic Sketchbook Series](#classic-sketchbook-series)
  - [Junas](#junas)
  - [Random Shapes](#random-shapes)
  - [Color Circle2](#color-circle2)
  - [Circle Grid](#circle-grid)
  - [Circle Composes Circle](#circle-composes-circle)
  - [Pixel Hole](#pixel-hole)
  - [Dots Wave](#dots-wave)
  - [Contour Line](#contour-line)
  - [Noise Line](#noise-line)
  - [Dot Line](#dot-line)
  - [Ocean Fish](#ocean-fish)
  - [Circle Loop](#circle-loop)
  - [Domain Warp](#domain-warp)
  - [Circle Noise](#circle-noise)
  - [Perlin Perls](#perlin-perls)
  - [Color Canva](#color-canva)
  - [Julia Set](#julia-set)
  - [Black Hole](#black-hole)
  - [Silk Sky](#silk-sky)
  - [Circle Move](#circle-move)
  - [Random Circle](#random-circle)

---

## Wave 27: Multi-Paradigm Non-Mœbius Flagship Quartet — Brazilian Agate & Amethyst Geode, Song Dynasty Blue-and-Green Shanshui Scroll, Imperial Fabergé Guilloché Enamel Star & 3D Polar Aurora Borealis Fjord

### Agate Geode Crystal

```go
package main

import (
	"image/color"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/arts"
	"github.com/jdxyw/generativeart/common"
)

func main() {
	c := generativeart.NewCanva(600, 600)
	c.SetSeed(20262701)
	c.SetBackground(color.RGBA{R: 7, G: 9, B: 18, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Dunhuang)
	c.Draw(arts.NewAgateGeodeCrystal(20, 3, 1.0))
	c.ToPNG("agategeodecrystal.png")
}
```

![Agate Geode Crystal](images/agategeodecrystal.png)

### Faberge Guilloche Ornament

```go
package main

import (
	"image/color"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/arts"
	"github.com/jdxyw/generativeart/common"
)

func main() {
	c := generativeart.NewCanva(600, 600)
	c.SetSeed(20262701)
	c.SetBackground(color.RGBA{R: 6, G: 8, B: 20, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Dunhuang)
	c.Draw(arts.NewFabergeGuillocheOrnament(12, 28, 1.0))
	c.ToPNG("fabergeguillocheornament.png")
}
```

![Faberge Guilloche Ornament](images/fabergeguillocheornament.png)

### Aurora Curtain Fold

```go
package main

import (
	"image/color"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/arts"
	"github.com/jdxyw/generativeart/common"
)

func main() {
	c := generativeart.NewCanva(600, 600)
	c.SetSeed(20262701)
	c.SetBackground(color.RGBA{R: 3, G: 6, B: 18, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Oceanic)
	c.Draw(arts.NewAuroraCurtainFold(4, 180, 1.15))
	c.ToPNG("auroracurtainfold.png")
}
```

![Aurora Curtain Fold](images/auroracurtainfold.png)

---

## Wave 26: Multi-Paradigm Non-Mœbius Flagship Quartet — Fidenza Velvet & Gold Loom, Kuroshio Bioluminescence LIC, 3D Lorentz Strange-Attractor Ribbons & Total Solar Eclipse Coronal Arcade

### Fidenza Velvet Loom

```go
package main

import (
	"image/color"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/arts"
	"github.com/jdxyw/generativeart/common"
)

func main() {
	c := generativeart.NewCanva(600, 600)
	c.SetSeed(20262601)
	c.SetBackground(color.RGBA{R: 7, G: 11, B: 22, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Dunhuang)
	c.Draw(arts.NewFidenzaVelvetLoom(18, 5, 2.2))
	c.ToPNG("fidenzavelvetloom.png")
}
```

![Fidenza Velvet Loom](images/fidenzavelvetloom.png)

### Kuroshio Bioluminescence

```go
package main

import (
	"image/color"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/arts"
	"github.com/jdxyw/generativeart/common"
)

func main() {
	c := generativeart.NewCanva(600, 600)
	c.SetSeed(20262601)
	c.SetBackground(color.RGBA{R: 4, G: 10, B: 28, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Oceanic)
	c.Draw(arts.NewKuroshioBioluminescence(14, 18, 24.0))
	c.ToPNG("kuroshiobioluminescence.png")
}
```

![Kuroshio Bioluminescence](images/kuroshiobioluminescence.png)

### Lorentz Strange Ribbon

```go
package main

import (
	"image/color"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/arts"
	"github.com/jdxyw/generativeart/common"
)

func main() {
	c := generativeart.NewCanva(600, 600)
	c.SetSeed(20262601)
	c.SetBackground(color.RGBA{R: 6, G: 9, B: 19, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Dunhuang)
	c.Draw(arts.NewLorentzStrangeRibbon(26, 28.0, 2.667))
	c.ToPNG("lorentzstrangeribbon.png")
}
```

![Lorentz Strange Ribbon](images/lorentzstrangeribbon.png)

### Coronal Magnetic Arcade

```go
package main

import (
	"image/color"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/arts"
	"github.com/jdxyw/generativeart/common"
)

func main() {
	c := generativeart.NewCanva(600, 600)
	c.SetSeed(20262601)
	c.SetBackground(color.RGBA{R: 5, G: 7, B: 16, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Inferno)
	c.Draw(arts.NewCoronalMagneticArcade(9, 19000, 1.15))
	c.ToPNG("coronalmagneticarcade.png")
}
```

![Coronal Magnetic Arcade](images/coronalmagneticarcade.png)

---

## Wave 25: Multi-Paradigm Non-Mœbius Flagship Quartet — 3D Hopf Fibration Silk Ribbons, 3D Laser-Cut Volcanic Atoll Terraces, 3D G-Buffer Peacock Ocelli Plumes & Cloisonné Star-Rose Mosaic

### Hopf Torus Silk

```go
package main

import (
	"image/color"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/arts"
	"github.com/jdxyw/generativeart/common"
)

func main() {
	c := generativeart.NewCanva(600, 600)
	c.SetSeed(20262501)
	c.SetBackground(color.RGBA{R: 6, G: 9, B: 19, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Dunhuang)
	c.Draw(arts.NewHopfTorusSilk(6, 22, 0.72))
	c.ToPNG("hopftorussilk.png")
}
```

![Hopf Torus Silk](images/hopftorussilk.png)

### Volcanic Atoll Terraces

```go
package main

import (
	"image/color"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/arts"
	"github.com/jdxyw/generativeart/common"
)

func main() {
	c := generativeart.NewCanva(600, 600)
	c.SetSeed(20262501)
	c.SetBackground(color.RGBA{R: 244, G: 236, B: 214, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Sunset)
	c.Draw(arts.NewVolcanicAtollTerraces(16, 4, 1.75))
	c.ToPNG("volcanicatollterraces.png")
}
```

![Volcanic Atoll Terraces](images/volcanicatollterraces.png)

### Peacock Ocelli Plumes

```go
package main

import (
	"image/color"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/arts"
	"github.com/jdxyw/generativeart/common"
)

func main() {
	c := generativeart.NewCanva(600, 600)
	c.SetSeed(20262501)
	c.SetBackground(color.RGBA{R: 12, G: 20, B: 30, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Oceanic)
	c.Draw(arts.NewPeacockOcelliPlumes(7, 18000, 390.0))
	c.ToPNG("peacockocelliplumes.png")
}
```

![Peacock Ocelli Plumes](images/peacockocelliplumes.png)

### Byzantine Rose Mosaic

```go
package main

import (
	"image/color"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/arts"
	"github.com/jdxyw/generativeart/common"
)

func main() {
	c := generativeart.NewCanva(600, 600)
	c.SetSeed(20262501)
	c.SetBackground(color.RGBA{R: 12, G: 14, B: 24, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Dunhuang)
	c.Draw(arts.NewByzantineRoseMosaic(12, 4, 1.15))
	c.ToPNG("byzantinerosemosaic.png")
}
```

![Byzantine Rose Mosaic](images/byzantinerosemosaic.png)

---

## Wave 24: Phase 2–5 Flagship Quartet — Venetian Murano Blown-Glass Nautilus, Haeckel 3D Silica Radiolaria, Relativistic Black-Hole Gravitational Lens & Weingarten Torus-Knot Sculpture

### Murano Glass Nautilus

```go
package main

import (
	"image/color"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/arts"
	"github.com/jdxyw/generativeart/common"
)

func main() {
	c := generativeart.NewCanva(600, 600)
	c.SetSeed(20262401)
	c.SetBackground(color.RGBA{R: 248, G: 241, B: 224, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Dunhuang)
	c.Draw(arts.NewMuranoGlassNautilus(14, 5, 0.048))
	c.ToPNG("muranoglassnautilus.png")
}
```

![Murano Glass Nautilus](images/muranoglassnautilus.png)

### Haeckel Radiolaria

```go
package main

import (
	"image/color"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/arts"
	"github.com/jdxyw/generativeart/common"
)

func main() {
	c := generativeart.NewCanva(600, 600)
	c.SetSeed(20262402)
	c.SetBackground(color.RGBA{R: 248, G: 241, B: 224, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Oceanic)
	c.Draw(arts.NewHaeckelRadiolaria(3, 6, 1.0))
	c.ToPNG("haeckelradiolaria.png")
}
```

![Haeckel Radiolaria](images/haeckelradiolaria.png)

### Relativistic Lens

```go
package main

import (
	"image/color"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/arts"
	"github.com/jdxyw/generativeart/common"
)

func main() {
	c := generativeart.NewCanva(600, 600)
	c.SetSeed(20262403)
	c.SetBackground(color.RGBA{R: 248, G: 241, B: 224, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Inferno)
	c.Draw(arts.NewRelativisticLens(4, 2, 0.85))
	c.ToPNG("relativisticlens.png")
}
```

![Relativistic Lens](images/relativisticlens.png)

### Weingarten Torus Knot

```go
package main

import (
	"image/color"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/arts"
	"github.com/jdxyw/generativeart/common"
)

func main() {
	c := generativeart.NewCanva(600, 600)
	c.SetSeed(20262404)
	c.SetBackground(color.RGBA{R: 248, G: 241, B: 224, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Dunhuang)
	c.Draw(arts.NewWeingartenTorusKnot(2, 3, 1.0))
	c.ToPNG("weingartentorusknot.png")
}
```

![Weingarten Torus Knot](images/weingartentorusknot.png)

---

## Wave 23: Master Architecture, Celestial Mechanics & Sci-Fi Megastructures — Borges Library of Babel, Kepler Mysterium Cosmographicum, Mœbius Hermetic Garage & Calabi-Yau Algebraic Manifold

### Babel Library

```go
package main

import (
	"image/color"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/arts"
	"github.com/jdxyw/generativeart/common"
)

func main() {
	c := generativeart.NewCanva(600, 600)
	c.SetSeed(20262301)
	c.SetBackground(color.RGBA{R: 248, G: 241, B: 224, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Dunhuang)
	c.Draw(arts.NewBabelLibrary(6, 5, 1.0))
	c.ToPNG("babellibrary.png")
}
```

![Babel Library](images/babellibrary.png)

### Kepler Harmonices

```go
package main

import (
	"image/color"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/arts"
	"github.com/jdxyw/generativeart/common"
)

func main() {
	c := generativeart.NewCanva(600, 600)
	c.SetSeed(20262302)
	c.SetBackground(color.RGBA{R: 248, G: 241, B: 224, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Dunhuang)
	c.Draw(arts.NewKeplerHarmonices(5, 3, 0.26))
	c.ToPNG("keplerharmonices.png")
}
```

![Kepler Harmonices](images/keplerharmonices.png)

### Moebius Hermetic Garage

```go
package main

import (
	"image/color"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/arts"
	"github.com/jdxyw/generativeart/common"
)

func main() {
	c := generativeart.NewCanva(600, 600)
	c.SetSeed(20262303)
	c.SetBackground(color.RGBA{R: 248, G: 241, B: 224, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Sunset)
	c.Draw(arts.NewMoebiusHermeticGarage(24, 6, 0.36))
	c.ToPNG("moebiushermeticgarage.png")
}
```

![Moebius Hermetic Garage](images/moebiushermeticgarage.png)

### Calabi Yau Manifold

```go
package main

import (
	"image/color"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/arts"
	"github.com/jdxyw/generativeart/common"
)

func main() {
	c := generativeart.NewCanva(600, 600)
	c.SetSeed(20262304)
	c.SetBackground(color.RGBA{R: 248, G: 241, B: 224, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Dunhuang)
	c.Draw(arts.NewCalabiYauManifold(5, 2, 0.35))
	c.ToPNG("calabiyaumanifold.png")
}
```

![Calabi Yau Manifold](images/calabiyaumanifold.png)

---

## Wave 22: Part VIII Flagship Quartet — Piranesi Carceri Vault Labyrinth, 3D SDF Mœbius Leviathan Airship, Dürer/Jamnitzer 3D Polyhedron & Kubelka-Munk Mineral Shan-Shui

### Piranesi Carceri

```go
package main

import (
	"image/color"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/arts"
	"github.com/jdxyw/generativeart/common"
)

func main() {
	c := generativeart.NewCanva(600, 600)
	c.SetSeed(20262201)
	c.SetBackground(color.RGBA{R: 248, G: 241, B: 224, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Dunhuang)
	c.Draw(arts.NewPiranesiCarceri(4, 4, 1.0))
	c.ToPNG("piranesicarceri.png")
}
```

![Piranesi Carceri](images/piranesicarceri.png)

### Moebius Leviathan

```go
package main

import (
	"image/color"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/arts"
	"github.com/jdxyw/generativeart/common"
)

func main() {
	c := generativeart.NewCanva(600, 600)
	c.SetSeed(20262202)
	c.SetBackground(color.RGBA{R: 248, G: 241, B: 224, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Sunset)
	c.Draw(arts.NewMoebiusLeviathan(14, 5, -0.18))
	c.ToPNG("moebiusleviathan.png")
}
```

![Moebius Leviathan](images/moebiusleviathan.png)

### Durer Polyhedron

```go
package main

import (
	"image/color"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/arts"
	"github.com/jdxyw/generativeart/common"
)

func main() {
	c := generativeart.NewCanva(600, 600)
	c.SetSeed(20262203)
	c.SetBackground(color.RGBA{R: 248, G: 241, B: 224, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Dunhuang)
	c.Draw(arts.NewDurerPolyhedron(12, 3, 0.32))
	c.ToPNG("durerpolyhedron.png")
}
```

![Durer Polyhedron](images/durerpolyhedron.png)

### Mineral Shan Shui

```go
package main

import (
	"image/color"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/arts"
	"github.com/jdxyw/generativeart/common"
)

func main() {
	c := generativeart.NewCanva(600, 600)
	c.SetSeed(20262204)
	c.SetBackground(color.RGBA{R: 244, G: 234, B: 210, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Dunhuang)
	c.Draw(arts.NewMineralShanShui(5, 4, 1.0))
	c.ToPNG("mineralshanshui.png")
}
```

![Mineral Shan Shui](images/mineralshanshui.png)

---

## Wave 21: Jean "Mœbius" Giraud Ligne Claire Quartet — Stratified Desert Canyon, L'Incal Megastructure Sanctuary, Gardens of Edena Travertine Oasis & Starwatcher Dyson-Torus Citadel

### Moebius Canyon

```go
package main

import (
	"image/color"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/arts"
	"github.com/jdxyw/generativeart/common"
)

func main() {
	c := generativeart.NewCanva(600, 600)
	c.SetSeed(20262101)
	c.SetBackground(color.RGBA{R: 248, G: 241, B: 224, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Dunhuang)
	c.Draw(arts.NewMoebiusCanyon(7, 18, 0.16))
	c.ToPNG("moebiuscanyon.png")
}
```

![Moebius Canyon](images/moebiuscanyon.png)

### Moebius Incal

```go
package main

import (
	"image/color"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/arts"
	"github.com/jdxyw/generativeart/common"
)

func main() {
	c := generativeart.NewCanva(600, 600)
	c.SetSeed(20262102)
	c.SetBackground(color.RGBA{R: 248, G: 241, B: 224, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Cyberpunk)
	c.Draw(arts.NewMoebiusIncal(8, 3, 1.0))
	c.ToPNG("moebiusincal.png")
}
```

![Moebius Incal](images/moebiusincal.png)

### Moebius Edena

```go
package main

import (
	"image/color"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/arts"
	"github.com/jdxyw/generativeart/common"
)

func main() {
	c := generativeart.NewCanva(600, 600)
	c.SetSeed(20262103)
	c.SetBackground(color.RGBA{R: 248, G: 241, B: 224, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Botanical)
	c.Draw(arts.NewMoebiusEdena(5, 16, 1.0))
	c.ToPNG("moebiusedena.png")
}
```

![Moebius Edena](images/moebiusedena.png)

### Moebius Starwatcher

```go
package main

import (
	"image/color"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/arts"
	"github.com/jdxyw/generativeart/common"
)

func main() {
	c := generativeart.NewCanva(600, 600)
	c.SetSeed(20262104)
	c.SetBackground(color.RGBA{R: 248, G: 241, B: 224, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Sunset)
	c.Draw(arts.NewMoebiusStarwatcher(3, 4, -0.24))
	c.ToPNG("moebiusstarwatcher.png")
}
```

![Moebius Starwatcher](images/moebiusstarwatcher.png)

---

## Wave 20: M.C. Escher Trio — Circle Limit Hyperbolic Tessellation, Print Gallery Droste Spiral & Isohedral Metamorphosis

### Escher Circle Limit

```go
package main

import (
	"image/color"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/arts"
	"github.com/jdxyw/generativeart/common"
)

func main() {
	c := generativeart.NewCanva(600, 600)
	c.SetSeed(20261101)
	c.SetBackground(color.RGBA{R: 243, G: 236, B: 224, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Dunhuang)
	c.Draw(arts.NewEscherCircleLimit(6, 4, 2))
	c.ToPNG("eschercirclelimit.png")
}
```

![Escher Circle Limit](images/eschercirclelimit.png)

### Escher Droste Gallery

```go
package main

import (
	"image/color"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/arts"
	"github.com/jdxyw/generativeart/common"
)

func main() {
	c := generativeart.NewCanva(600, 600)
	c.SetSeed(20261101)
	c.SetBackground(color.RGBA{R: 18, G: 22, B: 34, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Dunhuang)
	c.Draw(arts.NewEscherDrosteGallery(1, 8, 16.0))
	c.ToPNG("escherdrostegallery.png")
}
```

![Escher Droste Gallery](images/escherdrostegallery.png)

### Escher Metamorphosis

```go
package main

import (
	"image/color"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/arts"
	"github.com/jdxyw/generativeart/common"
)

func main() {
	c := generativeart.NewCanva(600, 600)
	c.SetSeed(20261101)
	c.SetBackground(color.RGBA{R: 248, G: 242, B: 228, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Dunhuang)
	c.Draw(arts.NewEscherMetamorphosis(10, 10, 1.0))
	c.ToPNG("eschermetamorphosis.png")
}
```

![Escher Metamorphosis](images/eschermetamorphosis.png)

---

## Wave 19: Analytical Perlin & Gabor Noise Quartet — Fidenza Flow Blocks, 3D Laser-Cut Topographic Terraces, 3D Perlin Silk Ribbon Sculpture & Celestial Gabor-Curl Astrolabe

### Perlin Malachite

```go
package main

import (
	"image/color"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/arts"
	"github.com/jdxyw/generativeart/common"
)

func main() {
	c := generativeart.NewCanva(600, 600)
	c.SetSeed(20261101)
	c.SetBackground(color.RGBA{R: 7, G: 24, B: 18, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Botanical)
	c.Draw(arts.NewPerlinMalachite(16, 5, 2.4))
	c.ToPNG("perlinmalachite.png")
}
```

![Perlin Malachite](images/perlinmalachite.png)

### Erosion Canyon

```go
package main

import (
	"image/color"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/arts"
	"github.com/jdxyw/generativeart/common"
)

func main() {
	c := generativeart.NewCanva(600, 600)
	c.SetSeed(20261101)
	c.SetBackground(color.RGBA{R: 12, G: 16, B: 28, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Sunset)
	c.Draw(arts.NewErosionCanyon(6, 1.8, 18.0))
	c.ToPNG("erosioncanyon.png")
}
```

![Erosion Canyon](images/erosioncanyon.png)

### Analytical Curl Silk

```go
package main

import (
	"image/color"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/arts"
	"github.com/jdxyw/generativeart/common"
)

func main() {
	c := generativeart.NewCanva(600, 600)
	c.SetSeed(20261101)
	c.SetBackground(color.RGBA{R: 10, G: 18, B: 40, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Dunhuang)
	c.Draw(arts.NewAnalyticalCurlSilk(420, 4, 2.2))
	c.ToPNG("analyticalcurlsilk.png")
}
```

![Analytical Curl Silk](images/analyticalcurlsilk.png)

### Gabor Wave Interference

```go
package main

import (
	"image/color"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/arts"
	"github.com/jdxyw/generativeart/common"
)

func main() {
	c := generativeart.NewCanva(600, 600)
	c.SetSeed(20261101)
	c.SetBackground(color.RGBA{R: 8, G: 12, B: 24, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Cyberpunk)
	c.Draw(arts.NewGaborWaveInterference(6, 3.2, 1.45))
	c.ToPNG("gaborwaveinterference.png")
}
```

![Gabor Wave Interference](images/gaborwaveinterference.png)

---

## Wave 18: LIC Jovian Magnetosphere, Thin-Film Iridescent Nacre, Kissing-Schottky Kleinian Pearls & 3D Bismuth Hopper Crystals

### Jupiter Magnetosphere

```go
package main

import (
	"image/color"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/arts"
	"github.com/jdxyw/generativeart/common"
)

func main() {
	c := generativeart.NewCanva(600, 600)
	c.SetSeed(20261030)
	c.SetBackground(color.RGBA{R: 9, G: 14, B: 28, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Oceanic)
	c.Draw(arts.NewJupiterMagnetosphere(9, 16, 22.0))
	c.ToPNG("jupitermagnetosphere.png")
}
```

![Jupiter Magnetosphere](images/jupitermagnetosphere.png)

### Iridescent Nacre

```go
package main

import (
	"image/color"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/arts"
	"github.com/jdxyw/generativeart/common"
)

func main() {
	c := generativeart.NewCanva(600, 600)
	c.SetSeed(20261030)
	c.SetBackground(color.RGBA{R: 14, G: 22, B: 30, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Oceanic)
	c.Draw(arts.NewIridescentNacre(7, 18000, 410.0))
	c.ToPNG("iridescentnacre.png")
}
```

![Iridescent Nacre](images/iridescentnacre.png)

### Schottky Kleinian

```go
package main

import (
	"image/color"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/arts"
	"github.com/jdxyw/generativeart/common"
)

func main() {
	c := generativeart.NewCanva(600, 600)
	c.SetSeed(20261030)
	c.SetBackground(color.RGBA{R: 7, G: 9, B: 20, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Dunhuang)
	c.Draw(arts.NewSchottkyKleinian(7, 1.0))
	c.ToPNG("schottkykleinian.png")
}
```

![Schottky Kleinian](images/schottkykleinian.png)

### Bismuth Hopper

```go
package main

import (
	"image/color"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/arts"
	"github.com/jdxyw/generativeart/common"
)

func main() {
	c := generativeart.NewCanva(600, 600)
	c.SetSeed(20261030)
	c.SetBackground(color.RGBA{R: 10, G: 12, B: 20, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Cyberpunk)
	c.Draw(arts.NewBismuthHopper(14, 9, 420.0))
	c.ToPNG("bismuthhopper.png")
}
```

![Bismuth Hopper](images/bismuthhopper.png)

---

## Wave 17: Tactile Bristle & Vortex Impasto Series — Binary Nebula Plumes, Abyssal Kuroshio Currents, Solar Prominence Tondo & Botanical Lotus Whorls

### Nebula Plumes

```go
package main

import (
	"image/color"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/arts"
	"github.com/jdxyw/generativeart/common"
)

func main() {
	c := generativeart.NewCanva(600, 600)
	c.SetSeed(20261029)
	c.SetBackground(color.RGBA{R: 243, G: 236, B: 225, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Dunhuang)
	c.Draw(arts.NewNebulaPlumes(3, 23000, 1.0))
	c.ToPNG("nebulaplumes.png")
}
```

![Nebula Plumes](images/nebulaplumes.png)

### Abyssal Currents

```go
package main

import (
	"image/color"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/arts"
	"github.com/jdxyw/generativeart/common"
)

func main() {
	c := generativeart.NewCanva(600, 600)
	c.SetSeed(20261029)
	c.SetBackground(color.RGBA{R: 13, G: 16, B: 24, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Cyberpunk)
	c.Draw(arts.NewAbyssalCurrents(18, 23000, 1.0))
	c.ToPNG("abyssalcurrents.png")
}
```

![Abyssal Currents](images/abyssalcurrents.png)

### Solar Prominence

```go
package main

import (
	"image/color"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/arts"
	"github.com/jdxyw/generativeart/common"
)

func main() {
	c := generativeart.NewCanva(600, 600)
	c.SetSeed(20261029)
	c.SetBackground(color.RGBA{R: 242, G: 235, B: 224, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Inferno)
	c.Draw(arts.NewSolarProminence(14, 24000, 1.0))
	c.ToPNG("solarprominence.png")
}
```

![Solar Prominence](images/solarprominence.png)

### Botanical Whorls

```go
package main

import (
	"image/color"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/arts"
	"github.com/jdxyw/generativeart/common"
)

func main() {
	c := generativeart.NewCanva(600, 600)
	c.SetSeed(20261029)
	c.SetBackground(color.RGBA{R: 238, G: 235, B: 226, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Botanical)
	c.Draw(arts.NewBotanicalWhorls(64, 22000, 137.507764))
	c.ToPNG("botanicalwhorls.png")
}
```

![Botanical Whorls](images/botanicalwhorls.png)

---

## Wave 16: Molnár Concentric Morph, Swirling Vortex Impasto, Mid-Century Sliced Pebbles, Hexagonal Mandala Honeycomb, Richter Squeegee Scrape, Retro Sci-Fi Halftone Panels & 3D Fluted Nautilus Spiral

### Concentric Morph

```go
package main

import (
	"image/color"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/arts"
	"github.com/jdxyw/generativeart/common"
)

func main() {
	c := generativeart.NewCanva(600, 600)
	c.SetSeed(20261028)
	c.SetBackground(color.RGBA{R: 22, G: 27, B: 34, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Dunhuang)
	c.Draw(arts.NewConcentricMorph(18, 18, 1.0))
	c.ToPNG("concentricmorph.png")
}
```

![Concentric Morph](images/concentricmorph.png)

### Vortex Impasto

```go
package main

import (
	"image/color"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/arts"
	"github.com/jdxyw/generativeart/common"
)

func main() {
	c := generativeart.NewCanva(600, 600)
	c.SetSeed(20261028)
	c.SetBackground(color.RGBA{R: 239, G: 232, B: 222, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Dunhuang)
	c.Draw(arts.NewVortexImpasto(46, 22000, 1.0))
	c.ToPNG("vorteximpasto.png")
}
```

![Vortex Impasto](images/vorteximpasto.png)

### Sliced Pebbles

```go
package main

import (
	"image/color"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/arts"
	"github.com/jdxyw/generativeart/common"
)

func main() {
	c := generativeart.NewCanva(600, 600)
	c.SetSeed(20261028)
	c.SetBackground(color.RGBA{R: 15, G: 18, B: 24, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Sunset)
	c.Draw(arts.NewSlicedPebbles(6, 3, 18.0))
	c.ToPNG("slicedpebbles.png")
}
```

![Sliced Pebbles](images/slicedpebbles.png)

### Hex Mandala

```go
package main

import (
	"image/color"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/arts"
	"github.com/jdxyw/generativeart/common"
)

func main() {
	c := generativeart.NewCanva(600, 600)
	c.SetSeed(20261028)
	c.SetBackground(color.RGBA{R: 7, G: 6, B: 5, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Dunhuang)
	c.Draw(arts.NewHexMandala(13, 16, 1.0))
	c.ToPNG("hexmandala.png")
}
```

![Hex Mandala](images/hexmandala.png)

### Squeegee Scrape

```go
package main

import (
	"image/color"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/arts"
	"github.com/jdxyw/generativeart/common"
)

func main() {
	c := generativeart.NewCanva(600, 600)
	c.SetSeed(20261028)
	c.SetBackground(color.RGBA{R: 24, G: 28, B: 32, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Sunset)
	c.Draw(arts.NewSqueegeeScrape(8, 460, 1.0))
	c.ToPNG("squeegeescrape.png")
}
```

![Squeegee Scrape](images/squeegeescrape.png)

### Halftone Panels

```go
package main

import (
	"image/color"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/arts"
	"github.com/jdxyw/generativeart/common"
)

func main() {
	c := generativeart.NewCanva(600, 600)
	c.SetSeed(20261028)
	c.SetBackground(color.RGBA{R: 244, G: 247, B: 244, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Cyberpunk)
	c.Draw(arts.NewHalftonePanels(22, 5.5))
	c.ToPNG("halftonepanels.png")
}
```

![Halftone Panels](images/halftonepanels.png)

### Nautilus Spiral (3D Fluted Ribbon Edition)

```go
package main

import (
	"image/color"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/arts"
	"github.com/jdxyw/generativeart/common"
)

func main() {
	c := generativeart.NewCanva(600, 600)
	c.SetSeed(20261028)
	c.SetBackground(color.RGBA{R: 18, G: 24, B: 34, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Dunhuang)
	c.Draw(arts.NewNautilusSpiral(5, 68))
	c.ToPNG("nautilusspiral.png")
}
```

![Nautilus Spiral](images/nautilusspiral.png)

---

## Wave 15: Hydraulic Erosion Alpine Peaks, Isometric Basalt Columns, Aeolian Sand Dunes, Bathymetric Trench Chart & 3D Wireframe Glacier Mesh

### Hydraulic Erosion

```go
package main

import (
	"image/color"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/arts"
	"github.com/jdxyw/generativeart/common"
)

func main() {
	c := generativeart.NewCanva(600, 600)
	c.SetSeed(20261027)
	c.SetBackground(color.RGBA{R: 10, G: 16, B: 34, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Oceanic)
	c.Draw(arts.NewHydraulicErosion(32000, 1.0))
	c.ToPNG("hydraulicerosion.png")
}
```

![Hydraulic Erosion](images/hydraulicerosion.png)

### Basalt Columns

```go
package main

import (
	"image/color"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/arts"
	"github.com/jdxyw/generativeart/common"
)

func main() {
	c := generativeart.NewCanva(600, 600)
	c.SetSeed(20261027)
	c.SetBackground(color.RGBA{R: 8, G: 11, B: 20, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Dunhuang)
	c.Draw(arts.NewBasaltColumns(16, 1.0))
	c.ToPNG("basaltcolumns.png")
}
```

![Basalt Columns](images/basaltcolumns.png)

### Sand Dunes

```go
package main

import (
	"image/color"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/arts"
	"github.com/jdxyw/generativeart/common"
)

func main() {
	c := generativeart.NewCanva(600, 600)
	c.SetSeed(20261027)
	c.SetBackground(color.RGBA{R: 12, G: 14, B: 32, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Sunset)
	c.Draw(arts.NewSandDunes(7, 1.0))
	c.ToPNG("sanddunes.png")
}
```

![Sand Dunes](images/sanddunes.png)

### Bathymetric Chart

```go
package main

import (
	"image/color"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/arts"
	"github.com/jdxyw/generativeart/common"
)

func main() {
	c := generativeart.NewCanva(600, 600)
	c.SetSeed(20261027)
	c.SetBackground(color.RGBA{R: 5, G: 10, B: 28, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Oceanic)
	c.Draw(arts.NewBathymetricChart(22, 7))
	c.ToPNG("bathymetricchart.png")
}
```

![Bathymetric Chart](images/bathymetricchart.png)

### Wireframe Terrain

```go
package main

import (
	"image/color"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/arts"
	"github.com/jdxyw/generativeart/common"
)

func main() {
	c := generativeart.NewCanva(600, 600)
	c.SetSeed(20261027)
	c.SetBackground(color.RGBA{R: 6, G: 9, B: 22, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Cyberpunk)
	c.Draw(arts.NewWireframeTerrain(76, 1.0))
	c.ToPNG("wireframeterrain.png")
}
```

![Wireframe Terrain](images/wireframeterrain.png)

---

## Wave 14: Riemann Zeta Critical Strip, 3D Hopf Fibration, Thomas Labyrinth Attractor, 12-Fold Faraday Quasicrystal Waves, Conformal Möbius Loxodrome & FTLE Lagrangian Coherent Structures

### Riemann Zeta

```go
package main

import (
	"image/color"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/arts"
	"github.com/jdxyw/generativeart/common"
)

func main() {
	c := generativeart.NewCanva(600, 600)
	c.SetSeed(20261027)
	c.SetBackground(color.RGBA{R: 6, G: 9, B: 18, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Dunhuang)
	c.Draw(arts.NewRiemannZeta(26, 16))
	c.ToPNG("riemannzeta.png")
}
```

![Riemann Zeta](images/riemannzeta.png)

### Hopf Fibration

```go
package main

import (
	"image/color"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/arts"
	"github.com/jdxyw/generativeart/common"
)

func main() {
	c := generativeart.NewCanva(600, 600)
	c.SetSeed(20261027)
	c.SetBackground(color.RGBA{R: 6, G: 8, B: 18, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Dunhuang)
	c.Draw(arts.NewHopfFibration(4, 24))
	c.ToPNG("hopffibration.png")
}
```

![Hopf Fibration](images/hopffibration.png)

### Thomas Attractor

```go
package main

import (
	"image/color"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/arts"
	"github.com/jdxyw/generativeart/common"
)

func main() {
	c := generativeart.NewCanva(600, 600)
	c.SetSeed(20261027)
	c.SetBackground(color.RGBA{R: 4, G: 7, B: 16, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Dunhuang)
	c.Draw(arts.NewThomasAttractor(900000, 0.19))
	c.ToPNG("thomasattractor.png")
}
```

![Thomas Attractor](images/thomasattractor.png)

### Faraday Waves

```go
package main

import (
	"image/color"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/arts"
	"github.com/jdxyw/generativeart/common"
)

func main() {
	c := generativeart.NewCanva(600, 600)
	c.SetSeed(20261027)
	c.SetBackground(color.RGBA{R: 7, G: 11, B: 24, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Dunhuang)
	c.Draw(arts.NewFaradayWaves(12, 1.0))
	c.ToPNG("faradaywaves.png")
}
```

![Faraday Waves](images/faradaywaves.png)

### Mobius Loxodrome

```go
package main

import (
	"image/color"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/arts"
	"github.com/jdxyw/generativeart/common"
)

func main() {
	c := generativeart.NewCanva(600, 600)
	c.SetSeed(20261027)
	c.SetBackground(color.RGBA{R: 7, G: 9, B: 18, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Dunhuang)
	c.Draw(arts.NewMobiusLoxodrome(14, 44))
	c.ToPNG("mobiusloxodrome.png")
}
```

![Mobius Loxodrome](images/mobiusloxodrome.png)

### Lagrangian Coherent

```go
package main

import (
	"image/color"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/arts"
	"github.com/jdxyw/generativeart/common"
)

func main() {
	c := generativeart.NewCanva(600, 600)
	c.SetSeed(20261027)
	c.SetBackground(color.RGBA{R: 5, G: 8, B: 18, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Cyberpunk)
	c.Draw(arts.NewLagrangianCoherent(18, 1.0))
	c.ToPNG("lagrangiancoherent.png")
}
```

![Lagrangian Coherent](images/lagrangiancoherent.png)

---

## Wave 13: 3D Halvorsen Attractor, Kleinian Fuchsian Pearls, Abelian Sandpile, Kuramoto Phase Defects, Farey Ford Circles & Quantum Chaos Scars (100-Art Milestone)

### Halvorsen Attractor

```go
package main

import (
	"image/color"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/arts"
	"github.com/jdxyw/generativeart/common"
)

func main() {
	c := generativeart.NewCanva(600, 600)
	c.SetSeed(20261023)
	c.SetBackground(color.RGBA{R: 4, G: 6, B: 14, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Cyberpunk)
	c.Draw(arts.NewHalvorsenAttractor(1200000, 1.4))
	c.ToPNG("halvorsenattractor.png")
}
```

![Halvorsen Attractor](images/halvorsenattractor.png)

### Hyperbolic Fuchsian

```go
package main

import (
	"image/color"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/arts"
	"github.com/jdxyw/generativeart/common"
)

func main() {
	c := generativeart.NewCanva(600, 600)
	c.SetSeed(20261023)
	c.SetBackground(color.RGBA{R: 7, G: 9, B: 18, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Dunhuang)
	c.Draw(arts.NewHyperbolicFuchsian(36, 1.0))
	c.ToPNG("hyperbolicfuchsian.png")
}
```

![Hyperbolic Fuchsian](images/hyperbolicfuchsian.png)

### Abelian Sandpile

```go
package main

import (
	"image/color"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/arts"
	"github.com/jdxyw/generativeart/common"
)

func main() {
	c := generativeart.NewCanva(600, 600)
	c.SetSeed(20261023)
	c.SetBackground(color.RGBA{R: 8, G: 14, B: 32, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Dunhuang)
	c.Draw(arts.NewAbelianSandpile(260000))
	c.ToPNG("abeliansandpile.png")
}
```

![Abelian Sandpile](images/abeliansandpile.png)

### Kuramoto Wave

```go
package main

import (
	"image/color"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/arts"
	"github.com/jdxyw/generativeart/common"
)

func main() {
	c := generativeart.NewCanva(600, 600)
	c.SetSeed(20261023)
	c.SetBackground(color.RGBA{R: 8, G: 11, B: 22, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Dunhuang)
	c.Draw(arts.NewKuramotoWave(8, 180))
	c.ToPNG("kuramotowave.png")
}
```

![Kuramoto Wave](images/kuramotowave.png)

### Ford Circles

```go
package main

import (
	"image/color"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/arts"
	"github.com/jdxyw/generativeart/common"
)

func main() {
	c := generativeart.NewCanva(600, 600)
	c.SetSeed(20261023)
	c.SetBackground(color.RGBA{R: 8, G: 10, B: 20, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Dunhuang)
	c.Draw(arts.NewFordCircles(14))
	c.ToPNG("fordcircles.png")
}
```

![Ford Circles](images/fordcircles.png)

### Quantum Billiard

```go
package main

import (
	"image/color"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/arts"
	"github.com/jdxyw/generativeart/common"
)

func main() {
	c := generativeart.NewCanva(600, 600)
	c.SetSeed(20261023)
	c.SetBackground(color.RGBA{R: 6, G: 9, B: 20, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Cyberpunk)
	c.Draw(arts.NewQuantumBilliard(48, 1.0))
	c.ToPNG("quantumbilliard.png")
}
```

![Quantum Billiard](images/quantumbilliard.png)

---

## Wave 12: Dunhuang Caisson, Celtic Knotwork, 3D Gyroid TPMS, Lichtenberg Plasma, Persian Muqarnas & Warring States Bronze

### Dunhuang Apsara

```go
func main() {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(600, 600)
	c.SetBackground(color.RGBA{R: 20, G: 14, B: 18, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Dunhuang)
	c.Draw(arts.NewDunhuangApsara(8, 5))
	c.ToPNG("dunhuangapsara.png")
}
```

![Dunhuang Apsara](images/dunhuangapsara.png)

### Celtic Knotwork

```go
func main() {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(600, 600)
	c.SetBackground(color.RGBA{R: 14, G: 18, B: 26, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Dunhuang)
	c.Draw(arts.NewCelticKnotwork(8, 1.0))
	c.ToPNG("celticknotwork.png")
}
```

![Celtic Knotwork](images/celticknotwork.png)

### Gyroid Surface

```go
func main() {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(600, 600)
	c.SetBackground(color.RGBA{R: 10, G: 14, B: 26, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Dunhuang)
	c.Draw(arts.NewGyroidSurface(4.2, 0.085))
	c.ToPNG("gyroidsurface.png")
}
```

![Gyroid Surface](images/gyroidsurface.png)

### Lichtenberg Tree

```go
func main() {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(600, 600)
	c.SetBackground(color.RGBA{R: 6, G: 8, B: 18, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Cyberpunk)
	c.Draw(arts.NewLichtenbergTree(9, 1.0))
	c.ToPNG("lichtenbergtree.png")
}
```

![Lichtenberg Tree](images/lichtenbergtree.png)

### Persian Tilework

```go
func main() {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(600, 600)
	c.SetBackground(color.RGBA{R: 16, G: 12, B: 24, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Dunhuang)
	c.Draw(arts.NewPersianTilework(16, 6))
	c.ToPNG("persiantilework.png")
}
```

![Persian Tilework](images/persiantilework.png)

### Champleve Enamel

```go
func main() {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(600, 600)
	c.SetBackground(color.RGBA{R: 12, G: 14, B: 18, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Dunhuang)
	c.Draw(arts.NewChampleveEnamel(8, 4))
	c.ToPNG("champleveenamel.png")
}
```

![Champleve Enamel](images/champleveenamel.png)

---

## Wave 11: Jovian Vortex Fluid, BZ Spiral Waves, Gilded PCB WFC, Cursive Calligraphy, Buddhabrot & Lyapunov

### Vortex Fluid

```go
func main() {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(600, 600)
	c.SetBackground(color.RGBA{R: 12, G: 22, B: 46, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Dunhuang)
	c.Draw(arts.NewVortexFluid(9, 18))
	c.ToPNG("vortexfluid.png")
}
```

![Vortex Fluid](images/vortexfluid.png)

### Cyclic Automata

```go
func main() {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(600, 600)
	c.SetBackground(color.RGBA{R: 10, G: 16, B: 34, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Sunset)
	c.Draw(arts.NewCyclicAutomata(14, 120))
	c.ToPNG("cyclicautomata.png")
}
```

![Cyclic Automata](images/cyclicautomata.png)

### Circuit WFC

```go
func main() {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(600, 600)
	c.SetBackground(color.RGBA{R: 8, G: 20, B: 26, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Cyberpunk)
	c.Draw(arts.NewCircuitWFC(28, 64))
	c.ToPNG("circuitwfc.png")
}
```

![Circuit WFC](images/circuitwfc.png)

### Asemic Calligraphy

```go
func main() {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(600, 600)
	c.SetBackground(color.RGBA{R: 244, G: 236, B: 214, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.UkiyoE)
	c.Draw(arts.NewAsemicCalligraphy(6, 5))
	c.ToPNG("asemiccalligraphy.png")
}
```

![Asemic Calligraphy](images/asemiccalligraphy.png)

### Buddhabrot

```go
func main() {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(600, 600)
	c.SetBackground(color.RGBA{R: 4, G: 6, B: 14, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Inferno)
	c.Draw(arts.NewBuddhabrot(450000, 360))
	c.ToPNG("buddhabrot.png")
}
```

![Buddhabrot](images/buddhabrot.png)

### Lyapunov Fractal

```go
func main() {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(600, 600)
	c.SetBackground(color.RGBA{R: 8, G: 12, B: 26, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Dunhuang)
	c.Draw(arts.NewLyapunovFractal("AABAB", 84))
	c.ToPNG("lyapunovfractal.png")
}
```

![Lyapunov Fractal](images/lyapunovfractal.png)

---

## Wave 10: 8-Fold Quasicrystals, Sashiko Embroidery, Electromagnetic Dipoles & Complex Newton Fractals

### Ammann Beenker

```go
func main() {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(600, 600)
	c.SetBackground(color.RGBA{R: 14, G: 14, B: 22, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Dunhuang)
	c.Draw(arts.NewAmmannBeenker(5))
	c.ToPNG("ammannbeenker.png")
}
```

![Ammann Beenker](images/ammannbeenker.png)

### Sashiko Weave

```go
func main() {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(600, 600)
	c.SetBackground(color.RGBA{R: 16, G: 28, B: 58, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.UkiyoE)
	c.Draw(arts.NewSashikoWeave(3))
	c.ToPNG("sashikoweave.png")
}
```

![Sashiko Weave](images/sashikoweave.png)

### Magnetic Dipole

```go
func main() {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(600, 600)
	c.SetBackground(color.RGBA{R: 8, G: 11, B: 22, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Cyberpunk)
	c.Draw(arts.NewMagneticDipole(6, 28))
	c.ToPNG("magneticdipole.png")
}
```

![Magnetic Dipole](images/magneticdipole.png)

### Newton Basin

```go
func main() {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(600, 600)
	c.SetBackground(color.RGBA{R: 10, G: 12, B: 22, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Dunhuang)
	c.Draw(arts.NewNewtonBasin(6, 48))
	c.ToPNG("newtonbasin.png")
}
```

![Newton Basin](images/newtonbasin.png)

---

## Wave 9: Topological Contours, Streamline Engraving, 3D Origami & Bauhaus SDF Relief

### Even Streamlines

```go
func main() {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(600, 600)
	c.SetBackground(color.RGBA{R: 12, G: 16, B: 28, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Oceanic)
	c.Draw(arts.NewEvenStreamlines(7.5, 3.9))
	c.ToPNG("evenstreamlines.png")
}
```

![Even Streamlines](images/evenstreamlines.png)

### Marching Isolines

```go
func main() {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(600, 600)
	c.SetBackground(color.RGBA{R: 12, G: 22, B: 42, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Dunhuang)
	c.Draw(arts.NewMarchingIsolines(16, 1.0))
	c.ToPNG("marchingisolines.png")
}
```

![Marching Isolines](images/marchingisolines.png)

### Origami Tessellation

```go
func main() {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(600, 600)
	c.SetBackground(color.RGBA{R: 16, G: 22, B: 44, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Dunhuang)
	c.Draw(arts.NewOrigamiTessellation(14, 14, 1.0))
	c.ToPNG("origamitessellation.png")
}
```

![Origami Tessellation](images/origamitessellation.png)

### Mondrian Subdivision

```go
func main() {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(600, 600)
	c.SetBackground(color.RGBA{R: 18, G: 18, B: 24, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Sunset)
	c.Draw(arts.NewMondrianSubdivision(4, 6.0))
	c.ToPNG("mondriansubdivision.png")
}
```

![Mondrian Subdivision](images/mondriansubdivision.png)

---

## Wave 8: Decorative Arts, Cloisonné Enamel & Printmaking

### Cloisonne Enamel

```go
func main() {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(600, 600)
	c.SetBackground(color.RGBA{R: 18, G: 14, B: 22, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Dunhuang)
	c.Draw(arts.NewCloisonneEnamel(6, 12))
	c.ToPNG("cloisonneenamel.png")
}
```

![Cloisonne Enamel](images/cloisonneenamel.png)

### Woodblock Wave

```go
func main() {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(600, 600)
	c.SetBackground(color.RGBA{R: 242, G: 232, B: 206, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.UkiyoE)
	c.Draw(arts.NewWoodblockWave(5, 42))
	c.ToPNG("woodblockwave.png")
}
```

![Woodblock Wave](images/woodblockwave.png)

### Art Deco Sunburst

```go
func main() {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(600, 600)
	c.SetBackground(color.RGBA{R: 14, G: 16, B: 24, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Sunset)
	c.Draw(arts.NewArtDecoSunburst(36, 7))
	c.ToPNG("artdecosunburst.png")
}
```

![Art Deco Sunburst](images/artdecosunburst.png)

### Stained Rose Window

```go
func main() {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(600, 600)
	c.SetBackground(color.RGBA{R: 12, G: 11, B: 18, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Cyberpunk)
	c.Draw(arts.NewStainedRoseWindow(12, 5))
	c.ToPNG("stainedrosewindow.png")
}
```

![Stained Rose Window](images/stainedrosewindow.png)

---

## Wave 7: Anisotropic Flow Mosaic & Low-Poly Crystal Series

### Mosaic Vortex

```go
func main() {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(600, 600)
	c.SetBackground(color.RGBA{R: 16, G: 20, B: 38, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Oceanic)
	c.Draw(arts.NewMosaicVortex(34, 1.0))
	c.ToPNG("mosaicvortex.png")
}
```

![Mosaic Vortex](images/mosaicvortex.png)

### Mosaic Wave

```go
func main() {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(600, 600)
	c.SetBackground(color.RGBA{R: 18, G: 24, B: 44, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.UkiyoE)
	c.Draw(arts.NewMosaicWave(32, 1.0))
	c.ToPNG("mosaicwave.png")
}
```

![Mosaic Wave](images/mosaicwave.png)

### Mosaic Canyon

```go
func main() {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(600, 600)
	c.SetBackground(color.RGBA{R: 22, G: 18, B: 32, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Dunhuang)
	c.Draw(arts.NewMosaicCanyon(32, 1.0))
	c.ToPNG("mosaiccanyon.png")
}
```

![Mosaic Canyon](images/mosaiccanyon.png)

### Mosaic Lotus

```go
func main() {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(600, 600)
	c.SetBackground(color.RGBA{R: 16, G: 24, B: 42, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Botanical)
	c.Draw(arts.NewMosaicLotus(12, 1.0))
	c.ToPNG("mosaiclotus.png")
}
```

![Mosaic Lotus](images/mosaiclotus.png)

### Mosaic Sun

```go
func main() {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(600, 600)
	c.SetBackground(color.RGBA{R: 22, G: 26, B: 48, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Sunset)
	c.Draw(arts.NewMosaicSun(32, 1.0))
	c.ToPNG("mosaicsun.png")
}
```

![Mosaic Sun](images/mosaicsun.png)

---

## Wave 6: G-Buffer Impasto, Voxel Terrain & Geometric Replications

### Impasto Flow

```go
func main() {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(600, 600)
	c.SetBackground(color.RGBA{R: 8, G: 7, B: 16, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Sunset)
	c.Draw(arts.NewImpastoFlow(3200, 520, 0.0032))
	c.ToPNG("impastoflow.png")
}
```

![Impasto Flow](images/impastoflow.png)

### Voxel Terrain

```go
func main() {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(600, 600)
	c.SetBackground(color.RGBA{R: 10, G: 12, B: 26, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Oceanic)
	c.Draw(arts.NewVoxelTerrain(340, 2.6, 1.0))
	c.ToPNG("voxelterrain.png")
}
```

![Voxel Terrain](images/voxelterrain.png)

### Packed Capsules

```go
func main() {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(600, 600)
	c.SetBackground(color.RGBA{R: 238, G: 236, B: 229, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Oceanic)
	c.Draw(arts.NewPackedCapsules(2200, 7.2, 3.6))
	c.ToPNG("packedcapsules.png")
}
```

![Packed Capsules](images/packedcapsules.png)

### Orphic Discs

```go
func main() {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(600, 600)
	c.SetBackground(color.RGBA{R: 72, G: 178, B: 102, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Youthful)
	c.Draw(arts.NewOrphicDiscs(34))
	c.ToPNG("orphicdiscs.png")
}
```

![Orphic Discs](images/orphicdiscs.png)

### Nautilus Spiral

```go
func main() {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(600, 600)
	c.SetBackground(color.RGBA{R: 18, G: 24, B: 34, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Dunhuang)
	c.Draw(arts.NewNautilusSpiral(5, 68))
	c.ToPNG("nautilusspiral.png")
}
```

![Nautilus Spiral](images/nautilusspiral.png)

### Tangent Cones

```go
func main() {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(600, 600)
	c.SetBackground(color.RGBA{R: 234, G: 88, B: 56, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Sunset)
	c.Draw(arts.NewTangentCones(26))
	c.ToPNG("tangentcones.png")
}
```

![Tangent Cones](images/tangentcones.png)

---

## Wave 5: Hyperbolic Geometry, Quasicrystals & 3D DoF Chaos

### Poincare Disk

```go
func main() {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(600, 600)
	c.SetBackground(color.RGBA{R: 8, G: 10, B: 20, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Dunhuang)
	c.Draw(arts.NewPoincareDisk(7, 3, 5))
	c.ToPNG("poincaredisk.png")
}
```

![Poincare Disk](images/poincaredisk.png)

### Suminagashi

```go
func main() {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(600, 600)
	c.SetBackground(color.RGBA{R: 244, G: 239, B: 229, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.UkiyoE)
	c.Draw(arts.NewSuminagashi(48, 7, 1.0))
	c.ToPNG("suminagashi.png")
}
```

![Suminagashi](images/suminagashi.png)

### Aizawa Attractor

```go
func main() {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(600, 600)
	c.SetBackground(color.RGBA{R: 5, G: 7, B: 15, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Cyberpunk)
	c.Draw(arts.NewAizawaAttractor(900000, 1.0))
	c.ToPNG("aizawaattractor.png")
}
```

![Aizawa Attractor](images/aizawaattractor.png)

### Penrose Tiling

```go
func main() {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(600, 600)
	c.SetBackground(color.RGBA{R: 10, G: 14, B: 24, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Oceanic)
	c.Draw(arts.NewPenroseTiling(6))
	c.ToPNG("penrosetiling.png")
}
```

![Penrose Tiling](images/penrosetiling.png)

### Isometric City

```go
func main() {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(600, 600)
	c.SetBackground(color.RGBA{R: 18, G: 20, B: 34, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Sunset)
	c.Draw(arts.NewIsometricCity(11, 1.0))
	c.ToPNG("isometriccity.png")
}
```

![Isometric City](images/isometriccity.png)

### DLA Frost

```go
func main() {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(600, 600)
	c.SetBackground(color.RGBA{R: 4, G: 8, B: 18, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Oceanic)
	c.Draw(arts.NewDLAFrost(18000, 0.55))
	c.ToPNG("dlafrost.png")
}
```

![DLA Frost](images/dlafrost.png)

---

## Wave 4: Fractal Flames, Flocking & Sacred Geometry

### Fractal Flame

```go
func main() {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(600, 600)
	c.SetBackground(color.RGBA{R: 4, G: 4, B: 10, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Magma)
	c.Draw(arts.NewFractalFlame(1800000, 5))
	c.ToPNG("fractalflame.png")
}
```

![Fractal Flame IFS](images/fractalflame.png)

### Joy Division

```go
func main() {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(600, 600)
	c.SetBackground(color.RGBA{R: 8, G: 10, B: 18, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Cyberpunk)
	c.Draw(arts.NewJoyDivision(72, 1.0, 1.0))
	c.ToPNG("joydivision.png")
}
```

![Joy Division Pulsar Waves](images/joydivision.png)

### Boids Flock

```go
func main() {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(600, 600)
	c.SetBackground(color.RGBA{R: 10, G: 14, B: 24, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Sunset)
	c.Draw(arts.NewBoidsFlock(240, 140, 52.0))
	c.ToPNG("boidsflock.png")
}
```

![Boids Murmuration Flock](images/boidsflock.png)

### Islamic Star

```go
func main() {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(600, 600)
	c.SetBackground(color.RGBA{R: 18, G: 22, B: 32, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Dunhuang)
	c.Draw(arts.NewIslamicStar(4, 67.5, 4.5))
	c.ToPNG("islamicstar.png")
}
```

![Islamic Girih Star Rosette](images/islamicstar.png)

### Delaunay Stained

```go
func main() {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(600, 600)
	c.SetBackground(color.RGBA{R: 12, G: 14, B: 22, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Plasma)
	c.Draw(arts.NewDelaunayStained(220, 0.65))
	c.ToPNG("delaunaystained.png")
}
```

![Delaunay Prismatic Stained Glass](images/delaunaystained.png)

### Phyllotaxis

```go
func main() {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(600, 600)
	c.SetBackground(color.RGBA{R: 12, G: 18, B: 20, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Botanical)
	c.Draw(arts.NewPhyllotaxis(420, 137.50776405, 11.5))
	c.ToPNG("phyllotaxis.png")
}
```

![Golden Angle Phyllotaxis Lotus](images/phyllotaxis.png)

---

## Wave 3: Reaction-Diffusion, Harmonographs & Venation

### Gray Scott

```go
func main() {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(600, 600)
	c.SetBackground(color.RGBA{R: 8, G: 12, B: 20, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Viridis)
	c.Draw(arts.NewGrayScott(3200, 0.037, 0.060))
	c.ToPNG("grayscott.png")
}
```

![Gray Scott Turing Coral](images/grayscott.png)

### Lissajous Ribbon

```go
func main() {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(600, 600)
	c.SetBackground(color.RGBA{R: 6, G: 8, B: 16, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Plasma)
	c.Draw(arts.NewLissajousRibbon(140, 3.0, 4.0, 5.0))
	c.ToPNG("lissajousribbon.png")
}
```

![Lissajous Silk Ribbon](images/lissajousribbon.png)

### Apollonian Gasket

```go
func main() {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(600, 600)
	c.SetBackground(color.RGBA{R: 10, G: 14, B: 24, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Dunhuang)
	c.Draw(arts.NewApollonianGasket(7, 2.5))
	c.ToPNG("apolloniangasket.png")
}
```

![Apollonian Gasket Astrolabe](images/apolloniangasket.png)

### Space Colonization

```go
func main() {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(600, 600)
	c.SetBackground(color.RGBA{R: 8, G: 14, B: 18, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Cyberpunk)
	c.Draw(arts.NewSpaceColonization(1800, 160, 6.0, 55.0))
	c.ToPNG("spacecolonization.png")
}
```

![Space Colonization Tree](images/spacecolonization.png)

### Truchet Flow

```go
func main() {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(600, 600)
	c.SetBackground(color.RGBA{R: 245, G: 240, B: 230, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.UkiyoE)
	c.Draw(arts.NewTruchetFlow(6, 3))
	c.ToPNG("truchetflow.png")
}
```

![Multi-Scale Truchet Flow](images/truchetflow.png)

### Aurora Wave

```go
func main() {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(600, 600)
	c.SetBackground(color.RGBA{R: 5, G: 8, B: 18, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Oceanic)
	c.Draw(arts.NewAuroraWave(4, 220, 1.0))
	c.ToPNG("aurorawave.png")
}
```

![Aurora Borealis Wave](images/aurorawave.png)

---

## Wave 2: Strange Attractors, Slime Molds & Shan-Shui Ink

### Clifford Nebula

```go
func main() {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(600, 600)
	c.SetBackground(color.RGBA{R: 6, G: 6, B: 14, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Inferno)
	c.Draw(arts.NewCliffordNebula(1500000, -1.7, 1.8, -1.9, -0.4))
	c.ToPNG("cliffordnebula.png")
}
```

![Clifford Nebula](images/cliffordnebula.png)

### Physarum

```go
func main() {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(600, 600)
	c.SetBackground(color.RGBA{R: 6, G: 10, B: 18, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Cyberpunk)
	c.Draw(arts.NewPhysarum(45000, 110, 0, 0))
	c.ToPNG("physarum.png")
}
```

![Physarum Network](images/physarum.png)

### Differential Growth

```go
func main() {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(600, 600)
	c.SetBackground(color.RGBA{R: 10, G: 12, B: 22, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Sunset)
	c.Draw(arts.NewDifferentialGrowth(220, 1600, 3.5, 7.5, 18.0))
	c.ToPNG("differentialgrowth.png")
}
```

![Differential Growth Coral](images/differentialgrowth.png)

### Chladni

```go
func main() {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(600, 600)
	c.SetBackground(color.RGBA{R: 8, G: 12, B: 22, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Oceanic)
	c.Draw(arts.NewChladni(90000, 7, 3, 1.0, -1.0))
	c.ToPNG("chladni.png")
}
```

![Chladni Cymatics](images/chladni.png)

### Voronoi Crystal

```go
func main() {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(600, 600)
	c.SetBackground(color.RGBA{R: 24, G: 22, B: 30, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Dunhuang)
	c.Draw(arts.NewVoronoiCrystal(85, 4))
	c.ToPNG("voronoicrystal.png")
}
```

![Voronoi Agate Crystal](images/voronoicrystal.png)

### Ink Landscape

```go
func main() {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(600, 600)
	c.SetBackground(color.RGBA{R: 244, G: 239, B: 228, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.UkiyoE)
	c.Draw(arts.NewInkLandscape(8, 0.45))
	c.ToPNG("inklandscape.png")
}
```

![Shan-Shui Ink Landscape](images/inklandscape.png)

---

## Wave 1: Curl Flow, Sand Painting, Watercolor & Guilloche

### Curl Flow

```go
func main() {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(600, 600)
	c.SetBackground(color.RGBA{R: 14, G: 18, B: 28, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Cyberpunk)
	c.Draw(arts.NewCurlFlow(110, 36, 0.0045))
	c.ToPNG("curlflow.png")
}
```

![Curl Flow](images/curlflow.png)

### Sand Ring

```go
func main() {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(600, 600)
	c.SetBackground(color.RGBA{R: 8, G: 10, B: 18, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Plasma)
	c.Draw(arts.NewSandRing(22, 22, 1.3))
	c.ToPNG("sandring.png")
}
```

![Sand Ring](images/sandring.png)

### Watercolor

```go
func main() {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(600, 600)
	c.SetBackground(color.RGBA{R: 247, G: 243, B: 234, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Dunhuang)
	c.SetLineColor(color.RGBA{R: 32, G: 28, B: 38, A: 255})
	c.Draw(arts.NewWatercolor(9, 22, 5, 0.32))
	c.ToPNG("watercolor.png")
}
```

![Watercolor](images/watercolor.png)

### Guilloche

```go
func main() {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(600, 600)
	c.SetBackground(color.RGBA{R: 12, G: 15, B: 26, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Oceanic)
	c.Draw(arts.NewGuilloche(28, 175, 65, 115))
	c.ToPNG("guilloche.png")
}
```

![Guilloche](images/guilloche.png)

---

## Classic Sketchbook Series

### Junas

```go
func main() {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(500, 500)
	c.SetBackground(common.Black)
	c.FillBackground()
	c.SetColorSchema(common.DarkRed)
	c.SetForeground(common.LightPink)
	c.Draw(arts.NewJanus(10, 0.2))
	c.ToPNG("janus.png")
}
```

![](images/janus.png)

### Random Shapes

```go
func main() {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(500, 500)
	c.SetBackground(common.White)
	c.FillBackground()
	c.SetColorSchema([]color.RGBA{
		{0xCF, 0x2B, 0x34, 0xFF},
		{0xF0, 0x8F, 0x46, 0xFF},
		{0xF0, 0xC1, 0x29, 0xFF},
		{0x19, 0x6E, 0x94, 0xFF},
		{0x35, 0x3A, 0x57, 0xFF},
	})
	c.Draw(arts.NewRandomShape(150))
	c.ToPNG("randomshape.png")
}
```

![](images/randomshape.png)

### Color Circle2

```go
func main() {
	rand.Seed(time.Now().Unix())
	colors := []color.RGBA{
		{0x11, 0x60, 0xC6, 0xFF},
		{0xFD, 0xD9, 0x00, 0xFF},
		{0xF5, 0xB4, 0xF8, 0xFF},
		{0xEF, 0x13, 0x55, 0xFF},
		{0xF4, 0x9F, 0x0A, 0xFF},
	}
	c := generativeart.NewCanva(800, 800)
	c.SetBackground(common.White)
	c.FillBackground()
	c.SetColorSchema(colors)
	c.Draw(arts.NewColorCircle2(30))
	c.ToPNG("colorcircle2.png")
}
```

![](images/colorcircle2.png)

### Circle Grid

```go
func main() {
	rand.Seed(time.Now().Unix())
	colors := []color.RGBA{
		{0xED, 0x34, 0x41, 0xFF},
		{0xFF, 0xD6, 0x30, 0xFF},
		{0x32, 0x9F, 0xE3, 0xFF},
		{0x15, 0x42, 0x96, 0xFF},
		{0x00, 0x00, 0x00, 0xFF},
		{0xFF, 0xFF, 0xFF, 0xFF},
	}
	c := generativeart.NewCanva(500, 500)
	c.SetBackground(color.RGBA{0xDF, 0xEB, 0xF5, 0xFF})
	c.FillBackground()
	c.SetColorSchema(colors)
	c.SetLineWidth(2.0)
	c.Draw(arts.NewCircleGrid(4, 6))
	c.ToPNG("circlegrid.png")
}
```

![](images/circlegrid.png)

### Circle Composes Circle

```go
func main() {
	rand.Seed(time.Now().Unix())
	colors := []color.RGBA{
		{0xF9, 0xC8, 0x0E, 0xFF},
		{0xF8, 0x66, 0x24, 0xFF},
		{0xEA, 0x35, 0x46, 0xFF},
		{0x66, 0x2E, 0x9B, 0xFF},
		{0x43, 0xBC, 0xCD, 0xFF},
	}
	c := generativeart.NewCanva(500, 500)
	c.SetBackground(color.RGBA{8, 10, 20, 255})
	c.FillBackground()
	c.SetColorSchema(colors)
	c.Draw(arts.NewCircleLoop2(7))
	c.ToPNG("colorloop2.png")
}
```

![](images/colorloop2.png)

### Pixel Hole

```go
func main() {
	rand.Seed(time.Now().Unix())
	colors := []color.RGBA{
		{0xF9, 0xC8, 0x0E, 0xFF},
		{0xF8, 0x66, 0x24, 0xFF},
		{0xEA, 0x35, 0x46, 0xFF},
		{0x66, 0x2E, 0x9B, 0xFF},
		{0x43, 0xBC, 0xCD, 0xFF},
	}
	c := generativeart.NewCanva(800, 800)
	c.SetBackground(common.Black)
	c.FillBackground()
	c.SetColorSchema(colors)
	c.SetIterations(1200)
	c.Draw(arts.NewPixelHole(60))
	c.ToPNG("pixelhole.png")
}
```

![](images/pixelhole.png)

### Dots Wave

```go
func main() {
	rand.Seed(time.Now().Unix())
	colors := []color.RGBA{
		{0xFF, 0xBE, 0x0B, 0xFF},
		{0xFB, 0x56, 0x07, 0xFF},
		{0xFF, 0x00, 0x6E, 0xFF},
		{0x83, 0x38, 0xEC, 0xFF},
		{0x3A, 0x86, 0xFF, 0xFF},
	}
	c := generativeart.NewCanva(500, 500)
	c.SetBackground(common.Black)
	c.FillBackground()
	c.SetColorSchema(colors)
	c.Draw(arts.NewDotsWave(300))
	c.ToPNG("dotswave.png")
}
```

![](images/dotswave.png)

### Contour Line

```go
func main() {
	rand.Seed(time.Now().Unix())
	colors := []color.RGBA{
		{0x58, 0x18, 0x45, 0xFF},
		{0x90, 0x0C, 0x3F, 0xFF},
		{0xC7, 0x00, 0x39, 0xFF},
		{0xFF, 0x57, 0x33, 0xFF},
		{0xFF, 0xC3, 0x0F, 0xFF},
	}
	c := generativeart.NewCanva(1600, 1600)
	c.SetBackground(color.RGBA{0x1a, 0x06, 0x33, 0xFF})
	c.FillBackground()
	c.SetColorSchema(colors)
	c.Draw(arts.NewContourLine(500))
	c.ToPNG("contourline.png")
}
```

![](images/contourline.png)

### Noise Line

```go
func main() {
	rand.Seed(time.Now().Unix())
	colors := []color.RGBA{
		{0x06, 0x7B, 0xC2, 0xFF},
		{0x84, 0xBC, 0xDA, 0xFF},
		{0xEC, 0xC3, 0x0B, 0xFF},
		{0xF3, 0x77, 0x48, 0xFF},
		{0xD5, 0x60, 0x62, 0xFF},
	}
	c := generativeart.NewCanva(1000, 1000)
	c.SetBackground(color.RGBA{0xF0, 0xFE, 0xFF, 0xFF})
	c.FillBackground()
	c.SetColorSchema(colors)
	c.Draw(arts.NewNoiseLine(1000))
	c.ToPNG("noiseline.png")
}
```

![](images/noiseline.png)

### Dot Line

```go
func main() {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(2080, 2080)
	c.SetBackground(color.RGBA{230, 230, 230, 255})
	c.SetLineWidth(10)
	c.SetIterations(15000)
	c.SetColorSchema(common.Plasma)
	c.FillBackground()
	c.Draw(arts.NewDotLine(100, 20, 50, false))
	c.ToPNG("dotline.png")
}
```

![](images/dotline.png)

### Ocean Fish

```go
func main() {
	rand.Seed(time.Now().Unix())
	colors := []color.RGBA{
		{0xCF, 0x2B, 0x34, 0xFF},
		{0xF0, 0x8F, 0x46, 0xFF},
		{0xF0, 0xC1, 0x29, 0xFF},
		{0x19, 0x6E, 0x94, 0xFF},
		{0x35, 0x3A, 0x57, 0xFF},
	}
	c := generativeart.NewCanva(500, 500)
	c.SetColorSchema(colors)
	c.Draw(arts.NewOceanFish(100, 8))
	c.ToPNG("oceanfish.png")
}
```

![](images/oceanfish.png)

### Circle Loop

```go
func main() {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(500, 500)
	c.SetBackground(common.Black)
	c.SetLineWidth(1)
	c.SetLineColor(common.Orange)
	c.SetAlpha(30)
	c.SetIterations(1000)
	c.FillBackground()
	c.Draw(arts.NewCircleLoop(100))
	c.ToPNG("circleloop.png")
}
```

![](images/circleloop.png)

### Domain Warp

```go
func cmap(r, m1, m2 float64) color.RGBA {
	rgb := color.RGBA{
		uint8(common.Constrain(m1*200*r, 0, 255)),
		uint8(common.Constrain(r*200, 0, 255)),
		uint8(common.Constrain(m2*255*r, 70, 255)),
		255,
	}
	return rgb
}

func main() {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(500, 500)
	c.SetBackground(common.Black)
	c.FillBackground()
	c.Draw(arts.NewDomainWrap(0.01, 4, 4, 20, cmap))
	c.ToPNG("domainwarp.png")
}
```

![](images/domainwarp.png)

### Circle Noise

```go
func main() {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(500, 500)
	c.SetBackground(common.White)
	c.SetAlpha(80)
	c.SetLineWidth(0.3)
	c.FillBackground()
	c.SetIterations(400)
	c.Draw(arts.NewCircleNoise(2000, 60, 80))
	c.ToPNG("circlenoise.png")
}
```

![](images/circlenoise.png)

### Perlin Perls

```go
func main() {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(500, 500)
	c.SetBackground(common.White)
	c.SetAlpha(120)
	c.SetLineWidth(0.3)
	c.FillBackground()
	c.SetIterations(200)
	c.Draw(arts.NewPerlinPerls(10, 200, 40, 80))
	c.ToPNG("perlinperls.png")
}
```

![](images/perlinperls.png)

### Color Canva

```go
func main() {
	rand.Seed(time.Now().Unix())
	colors := []color.RGBA{
		{0xF9, 0xC8, 0x0E, 0xFF},
		{0xF8, 0x66, 0x24, 0xFF},
		{0xEA, 0x35, 0x46, 0xFF},
		{0x66, 0x2E, 0x9B, 0xFF},
		{0x43, 0xBC, 0xCD, 0xFF},
	}
	c := generativeart.NewCanva(500, 500)
	c.SetBackground(common.Black)
	c.FillBackground()
	c.SetLineWidth(8)
	c.SetColorSchema(colors)
	c.Draw(arts.NewColorCanve(5))
	c.ToPNG("colorcanva.png")
}
```

![](images/colorcanva.png)

### Julia Set

```go
func julia1(z complex128) complex128 {
	c := complex(-0.1, 0.651)
	z = z*z + c
	return z
}

func main() {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(500, 500)
	c.SetIterations(800)
	c.SetColorSchema(common.Viridis)
	c.FillBackground()
	c.Draw(arts.NewJulia(julia1, 40, 1.5, 1.5))
	c.ToPNG("julia.png")
}
```

![](images/julia.png)

### Black Hole

```go
func main() {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(500, 500)
	c.SetBackground(color.RGBA{R: 30, G: 30, B: 30, A: 255})
	c.FillBackground()
	c.SetLineWidth(1.0)
	c.SetLineColor(common.Tomato)
	c.Draw(arts.NewBlackHole(200, 400, 0.01))
	c.ToPNG("blackhole.png")
}
```

![](images/blackhole.png)

### Silk Sky

```go
func main() {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(600, 600)
	c.SetAlpha(10)
	c.Draw(arts.NewSilkSky(15, 5))
	c.ToPNG("silksky.png")
}
```

![](images/silksky.png)

### Circle Move

```go
func main() {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(1200, 500)
	c.SetBackground(common.White)
	c.FillBackground()
	c.Draw(arts.NewCircleMove(1000))
	c.ToPNG("circlemove.png")
}
```

![](images/circlemove.png)

### Random Circle

```go
func main() {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(500, 500)
	c.SetBackground(common.MistyRose)
	c.SetLineWidth(1.0)
	c.SetLineColor(color.RGBA{
		R: 122,
		G: 122,
		B: 122,
		A: 30,
	})
	c.SetColorSchema(common.Plasma)
	c.SetIterations(4)
	c.FillBackground()
	c.Draw(arts.NewRandCicle(30, 80, 0.2, 2, 10, 30, true))
	c.ToPNG("randcircle.png")
}
```

![](images/randcircle.png)
