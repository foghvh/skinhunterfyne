import React, { useState, useEffect, useRef, memo } from "react";
import {
  Card,
  Box,
  Text,
  Dialog,
  Flex,
  Button,
  Skeleton,
  Inset,
  ScrollArea,
  Popover,
  Callout,
  IconButton,
  Separator,
  HoverCard,
  Link,
  Avatar,
  Heading,
  Tabs,
  RadioCards,


} from "@radix-ui/themes";

import { SkinItem } from "../pages/champions";
import { InfoCircledIcon, Cross2Icon, CircleBackslashIcon, DownloadIcon, AccessibilityIcon } from "@radix-ui/react-icons";
import { KhadaUrl } from "../data/data";
import { useUser } from "../context/usercontext";
import { useDownloadSkin } from "../context/download";
import { GetInstalledSkins } from "../../wailsjs/go/main/App";
// Add this image cache to prevent reloading the same images
const imageCache = new Map();

// Create a memory-efficient dialog component
const MemoryEfficientDialog = memo(({ children, isOpen, onClose }) => {
  useEffect(() => {
    // Clean up when dialog closes
    return () => {
      if (!isOpen) {
        // Force garbage collection when dialog closes
        setTimeout(() => {
          if (window.gc) window.gc();
        }, 100);
      }
    };
  }, [isOpen]);
  
  return children;
});

// Elimina imageCache y MemoryEfficientDialog

const SkinsGrid = ({ skins, champKey, asset, rarity, chromas }) => {
  const [visibleSkins, setVisibleSkins] = useState([]);
  const allSkinsRef = useRef([]);
  const gridRef = useRef(null);

  useEffect(() => {
    allSkinsRef.current = skins.filter(skin => !skin.isBase);
    setVisibleSkins(allSkinsRef.current.slice(0, 8));

    const observer = new IntersectionObserver((entries) => {
      if (entries[0].isIntersecting) {
        setVisibleSkins(prev => {
          if (prev.length >= allSkinsRef.current.length) return prev;
          return [...prev, ...allSkinsRef.current.slice(prev.length, prev.length + 4)];
        });
      }
    }, { rootMargin: '500px' });

    if (gridRef.current) {
      observer.observe(gridRef.current);
    }

    return () => {
      if (observer) observer.disconnect();
      allSkinsRef.current = [];
      setVisibleSkins([]);
    };
  }, [skins]);

  return (
    <Box
      ref={gridRef}
      className="grid grid-cols-[repeat(auto-fit,250px)] gap-3 justify-center items-center pb-7 justify-items-center"
    >
      {visibleSkins.map((skin) => (
        <Dialog.Root key={skin.id}>
          <Dialog.Trigger asChild>
            <div className="cursor-pointer w-fit h-fit">
              <SkinItem
                skin={skin}
                champKey={champKey}
                asset={asset}
                rarity={rarity}
              />
            </div>
          </Dialog.Trigger>
          <Dialog.Content>
            <SkinDialog
              skin={skin}
              champKey={champKey}
              asset={asset}
              rarity={rarity}
              chromas={chromas}
            />
          </Dialog.Content>
        </Dialog.Root>
      ))}

      {visibleSkins.length < allSkinsRef.current.length && (
        <div className="col-span-full flex justify-center py-4">
          <Skeleton>
            <Box style={{ width: "200px", height: "10px" }} />
          </Skeleton>
        </div>
      )}
    </Box>
  );
};

// Optimize SkinDialog component
const SkinDialog = memo(({ skin, champKey, asset, rarity, chromas }) => {
  const [isLoadingDs, setIsLoadingDs] = useState(true);
  const [isInstalled, setIsInstalled] = useState(false);
  const { userData, setUserData } = useUser();
  const downloadSkin = useDownloadSkin();
  const handleImageLoads = () => setIsLoadingDs(false);
  const filteredChromas = chromas.filter(chroma => chroma.origin === skin.id);
  const [selectedChroma, setSelectedChroma] = useState(null);

  useEffect(() => {
    async function checkInstallation() {
      try {
        const installedSkins = await GetInstalledSkins();
        if (!installedSkins) {
          setIsInstalled(false);
          return;
        }
        if (!Array.isArray(installedSkins)) {
          console.error("Installed skins data is not an array:", installedSkins);
          setIsInstalled(false);
          return;
        }
        setIsInstalled(installedSkins.some(installedSkin => 
          installedSkin.championId === String(Math.floor(skin.id / 1000))
        ));
      } catch (error) {
        console.error("Error checking installation:", error);
      }
    }
    checkInstallation();
  }, [skin.id]);


  
  const handleSelectChroma = (value) => {
    setSelectedChroma((prev) => (prev === value ? null : value));
  };

  // In the handleDownload function of SkinDialog component
  const handleDownload = async () => {
    const idToDownload = selectedChroma ? parseInt(selectedChroma) : skin.id;
  
    const selectedChromaData = selectedChroma
      ? chromas.find(chroma => String(chroma.id) === selectedChroma)
      : null;
  
    // Update global status
    if (window.updateGlobalStatus) {
      window.updateGlobalStatus(`Installing ${selectedChromaData ? 'chroma' : 'skin'}: ${skin.name}`);
    }
  
    await downloadSkin(
      Math.floor(skin.id / 1000),
      idToDownload,
      setUserData,
      skin,
      selectedChromaData
    );
  
    // Update global status after download
    if (window.updateGlobalStatus) {
      window.updateGlobalStatus(`Installed ${selectedChromaData ? 'chroma' : 'skin'}: ${skin.name}`);
    }
  };
  
  // Similarly in handleUninstall
  const handleUninstall = async () => {
    const loadingToast = toast.loading('Uninstalling skin...');
    try {
      const championId = Math.floor(skin.id / 1000);
      
      // Update global status
      if (window.updateGlobalStatus) {
        window.updateGlobalStatus(`Uninstalling skin: ${skin.name}`);
      }
      
      const result = await window.modTools.uninstallSkin(championId);
      
      if (!result.success) {
        throw new Error(result.error || 'Failed to uninstall skin');
      }
  
      setIsInstalled(false);
      toast.dismiss(loadingToast);
      toast.success('Skin uninstalled successfully');
  
      // Update global status after uninstall
      if (window.updateGlobalStatus) {
        window.updateGlobalStatus(`Uninstalled skin: ${skin.name}`);
      }
    } catch (error) {
      toast.dismiss(loadingToast);
      toast.error(error.message || 'Failed to uninstall skin');
    }
  };

  

  return (
    <>





      <Dialog.Content className="custom-dialog" maxWidth="850px" aria-describedby={undefined}>
        <Dialog.Title>              <Flex direction="row" align="center" mt="2">
          {rarity(skin) && (
            <img
              className="mr-2 "
              width="20"
              height="auto"
              src={rarity(skin)[0]}
              title={rarity(skin)[1]}
              alt={rarity(skin)[1]}
            />
          )}

          {skin.name}</Flex></Dialog.Title>

        <Flex
          direction={{
            md: "row",
            xl: "row",
          }}
          justify={{
            md: "space-between",
            xl: "space-between",
          }}
          align={{
            md: "start",
            xl: "start-between",
          }}
          gap="4"
          className="w-full"
          style={{ transition: 'all 1s ease-in-out' }}
        >
          {/* Skin Image and Bio */}
          <Box style={{ flex: 1, maxWidth: "100%" }}>
            <Card
              size="4"
              style={{
                width: "100%",
                maxWidth: "450px",
                padding: "1rem",
              }}
            >


              <Inset
                clip="border-box"
                side="all"
                pb="current"
                size=""
              >
                {isLoadingDs && (
                  <Box
                    className="skeleton-container bg-primary"
                    style={{
                      width: '-webkit-fill-available',
                      objectFit: "contain",
                      height: '1000px',
                      minHeight: 'auto',
                      minWidth: 'auto',
                      maxHeight: '270px'
                    }}
                  >
                    <Skeleton className="w-full h-full" />
                  </Box>
                )}
                <img
                  src={skin?.splashPath ? asset(skin.splashPath) : ""}
                  alt="Champion Splash"
                  style={{
                    display: isLoadingDs ? "none" : "block",
                    objectFit: "cover",
                    width: "-webkit-fill-available",
                    height: "270px",
                    maxHeight: "300px",
                  }}
                  onLoad={handleImageLoads}
                />
              </Inset>

              <ScrollArea
                type="auto"
                scrollbars="vertical"
                style={{
                  height: skin.description ? "auto" : "auto",
                  maxHeight: skin.description ? "130px" : "auto",
                  minHeight: "30px",

                  verticalAlign: skin.description ? "none" : "center"
                }}
              >
                <Text
                  as="p"
                  size="3"
                  className="pr-4 pt-2"
                  style={{
                    textAlign: skin.description ? "left" : "left",
                  }}
                >
                  {skin.description || "This skin does not have a description."}
                </Text>
              </ScrollArea>

            </Card>
            <Callout.Root variant="soft" my="3" size="1" color="yellow">
                      <Callout.Icon>
                        <AccessibilityIcon />
                      </Callout.Icon>
                      <Callout.Text>
                      This skin may not work properly due to game updates
                      </Callout.Text>
                    </Callout.Root>
          </Box>

          <Flex direction="column" gap="3" align="start" className="w-full" style={{ flex: 1 }}>
            {/* Difficulty */}
            <Flex direction="column" style={{ width: "100%" }}>
              <Text as="h1" size="5" mb="4">View skin on{" "}<HoverCard.Root>
                <HoverCard.Trigger>
                  <Link href={KhadaUrl(skin.id, selectedChroma)} target="_blank">
                    Model viewer
                  </Link>
                </HoverCard.Trigger>
                <HoverCard.Content maxWidth="300px">
                  <Flex gap="4">
                    <Avatar
                      size="3"
                      fallback="R"
                      radius="full"
                      src="https://modelviewer.lol/logo.svg"
                    />
                    <Box>
                      <Heading size="3" as="h3">
                        Khada
                      </Heading>
                      <Text as="div" size="2" color="gray" mb="2">
                        Model Viewer for LoL
                      </Text>
                      <Text as="div" size="2">
                        Fan-made 3D model viewer for every champion and skin from League of Legends with chromas and alternative forms!
                      </Text>
                    </Box>
                  </Flex>
                </HoverCard.Content>
              </HoverCard.Root>
                <Popover.Root>
                  <Popover.Trigger>
                    <IconButton variant="soft" ml="2" size="1">
                      <InfoCircledIcon />
                    </IconButton>
                  </Popover.Trigger>
                  <Popover.Content size="3" maxWidth="300px">
                    <Text as="p" trim="both" size="3">
                      Preview the in-game appearance of the skin.
                    </Text>
                  </Popover.Content>
                </Popover.Root></Text>


              <Separator size="4" orientation="horizontal" />
            </Flex>

            <Flex direction="column" style={{ width: "100%" }}>
              <Text as="h1" size="5">Chromas
                <Popover.Root>
                  <Popover.Trigger>
                    <IconButton ml="2" variant="soft" size="1">
                      <InfoCircledIcon />
                    </IconButton>
                  </Popover.Trigger>
                  <Popover.Content size="3" maxWidth="300px">
                    <Text as="p" trim="both" size="3">
                    Change skin colors
                    </Text>
                  </Popover.Content>
                </Popover.Root>
              </Text>
              <Tabs.Root defaultValue="circles">
                <Tabs.List>
                  <Tabs.Trigger value="circles">Circles</Tabs.Trigger>
                  <Tabs.Trigger value="images">Images</Tabs.Trigger>

                </Tabs.List>

                <Box pt="3">
                  <Tabs.Content value="circles">
                    <Text size="2">Select and Download your skin.</Text>
                    <ScrollArea type="auto" className="max-h-64 w-full border border-primary rounded-lg p-2">
                      <RadioCards.Root value={selectedChroma} onValueChange={handleSelectChroma} className="grid md:grid-flow-row md:grid-cols-[repeat(auto-fit,minmax(100px,1fr))] md:place-items-center grid-cols-[repeat(auto-fit,minmax(100px,1fr))] place-items-center gap-2">

                        {/* Opción Default */}
                        <RadioCards.Item
                          key="default"
                          value={String(skin.id)}
                          className={`w-24 h-28 p-1 border border-gray-300 rounded-lg shadow-md hover:shadow-lg transition-shadow flex flex-col items-center justify-center
                        ${selectedChroma === String(skin.id) ? "ring-2 ring-secondary" : ""}`}
                        >
                          <CircleBackslashIcon className="w-8 h-8 text-gray-500" />
                          <Text align="center" className="mt-2 text-sm font-medium">Default</Text>
                        </RadioCards.Item>

                        {/* Chromas map */}
                        {filteredChromas.length > 0 ? (
                          filteredChromas.map((chroma) => {
                            const colors = chroma.colors || ["#000000"]; 
                            const gradient = colors.length > 1
                              ? `linear-gradient(to right bottom, ${colors[0]} 50%, ${colors[1]} 50.3%)`
                              : colors[0]; 
                            return (
                              <RadioCards.Item
                                key={chroma.id}
                                value={String(chroma.id)}
                                className={`w-24 h-28 p-1 border border-gray-300 rounded-lg shadow-md hover:shadow-lg transition-shadow flex flex-col items-center justify-center
              ${selectedChroma === String(chroma.id) ? "ring-2 ring-secondary" : ""}`}
                              >
                                <div
                                  className="w-16 h-16 rounded-full border border-gray-400"
                                  style={{ background: gradient }}
                                />

                                <Text align="center" className="mt-2 text-sm font-medium">{chroma.name}</Text>
                              </RadioCards.Item>
                            );
                          })
                        ) : (
                          <></>
                        )}
                      </RadioCards.Root>
                    </ScrollArea>

                  </Tabs.Content>

                  <Tabs.Content value="images">
                    <Text size="2">Select and Download your skin.</Text>

                    <ScrollArea type="auto" className="max-h-64 w-full border border-primary rounded-lg p-2">
                      <RadioCards.Root value={selectedChroma} onValueChange={handleSelectChroma}  className="grid md:grid-flow-row md:grid-cols-[repeat(auto-fit,minmax(100px,1fr))] md:place-items-center grid-cols-[repeat(auto-fit,minmax(100px,1fr))] place-items-center gap-2">

                        <RadioCards.Item
                          key="default"
                          value={String(skin.id)}
                          className={`w-24 h-28 p-1 border border-gray-300 rounded-lg shadow-md hover:shadow-lg transition-shadow flex flex-col items-center justify-center
        ${selectedChroma === String(skin.id) ? "ring-2 ring-secondary" : ""}`}
                        >
                          <CircleBackslashIcon className="w-8 h-8 text-gray-500" />
                          <Text align="center" className="mt-2 text-sm font-medium">Default</Text>
                        </RadioCards.Item>

                        {filteredChromas.length > 0 ? (
                          filteredChromas.map((chroma) => (
                            <RadioCards.Item
                              key={chroma.id}
                              value={String(chroma.id)}
                              className={`w-24 h-28 p-1 border border-gray-300 rounded-lg shadow-md hover:shadow-lg transition-shadow flex flex-col items-center justify-center
            ${selectedChroma === String(chroma.id) ? "ring-2 ring-secondary" : ""}`}
                            >
                              <img
                                className="w-16 h-auto rounded-md"
                                src={chroma.chromaPath}
                                alt={chroma.name}
                                onLoad={handleImageLoads}
                              />
                              <Text align="center" className="mt-2 text-sm font-medium">{chroma.name}</Text>
                            </RadioCards.Item>
                          ))
                        ) : (
                          <></>
                        )}
                      </RadioCards.Root>
                    </ScrollArea>
                    
                  </Tabs.Content>

                  <Callout.Root variant="surface" my="3">
                      <Callout.Icon>
                        <InfoCircledIcon />
                      </Callout.Icon>
                      <Callout.Text>
                        This is going to consume a credit {userData?.fichasporskin > 0 && ` (${userData.fichasporskin} credits left)`}
                      </Callout.Text>
                    </Callout.Root>
                </Box>
              </Tabs.Root>

            </Flex>

          </Flex>

        </Flex>





        {/* Close button */}
        <Flex className="relative justify-between">
        <Dialog.Close asChild>
          <Button size="3" className="m-3 relative" variant="soft" color="gray">
            Close
          </Button>
        </Dialog.Close>
        
        {isInstalled ? (
          <Button 
            onClick={handleUninstall}
            disabled={!userData || userData.fichasporskin <= 0 || isInstalled}

            size="3" 
            color="red" 
            className="m-3 relative"
          >
            <Cross2Icon />
            Uninstall Skin
          </Button>
        ) : (
          <Button 
            onClick={handleDownload}
            disabled={!userData || userData.fichasporskin <= 0}
            size="3" 
            color="gray" 
            className="m-3 relative"
          >
            <DownloadIcon />
            Download {selectedChroma ? 'Chroma' : 'Skin'}
          </Button>
        )}
      </Flex>
      </Dialog.Content>
    </>
  );
}
);

export { SkinsGrid, SkinDialog };
