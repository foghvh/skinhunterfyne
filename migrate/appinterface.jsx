import React, { useState, useEffect } from 'react';
import { ArrowLeft, Search, Settings, Grid, Download, User, Layers } from "lucide-react";
import { ChampionsIndex, Champion } from './champions';
import { Omnisearch } from '../components/Omnisearch';
import { useParams } from 'react-router-dom';
import InstalledSkinsDialog from '../components/InstalledSkin';
import ContainerForm from '../components/ContainerForm';
import { Code, Container, Flex, Heading, Text, Section, Dialog, ScrollArea } from '@radix-ui/themes';
import { ProfileDialog } from '../components/ProfileDialog';
import { useUser } from '../context/usercontext';
import { SkinLine, SkinLinesIndex } from '../components/SkinLines';
import { motion, AnimatePresence } from 'framer-motion';
import ModOverlayButton from '../components/ModOverlayButton';
import { GetUserData, GetModStatus } from '../../wailsjs/go/main/App';
import { EventsOn, EventsOff } from "../../wailsjs/runtime/runtime";
// Import virtualization and lazy loading components
import { FixedSizeGrid } from 'react-window';
import AutoSizer from 'react-virtualized-auto-sizer';
import { LazyLoadImage } from 'react-lazy-load-image-component';
import 'react-lazy-load-image-component/src/effects/blur.css';

const TabTransition = ({ children, className }) => (
  <motion.div
    initial={{ opacity: 0, y: 20 }}
    animate={{ opacity: 1, y: 0 }}
    exit={{ opacity: 0, y: -20 }}
    transition={{ duration: 0.3, ease: "easeInOut" }}
    className={className}
  >
    {children}
  </motion.div>
);

const IconToggle = ({ showingSkinLines }) => (
  <motion.img
    initial={{ scale: 0.8, opacity: 0 }}
    animate={{ scale: 1, opacity: 1 }}
    exit={{ scale: 0.8, opacity: 0 }}
    transition={{ duration: 0.2 }}
    src={showingSkinLines
      ? "https://raw.communitydragon.org/latest/plugins/rcp-fe-lol-static-assets/global/default/images/skin-viewer/icon-three-masks-default.svg"
      : "https://raw.communitydragon.org/latest/plugins/rcp-fe-lol-static-assets/global/default/images/reward-tracker/item-category-icons/champion.png"
    }
    alt={showingSkinLines ? "Skin Lines" : "Champions"}
    className="w-6 h-6 filter brightness-75 hover:brightness-100 transition-all grayscale"
  />
);

// Modify the ChampionsIndex component to use virtualization
// Add this import at the top of the file
import { EnhancedImage } from './champions';

// Then in the VirtualizedChampionsIndex component, replace LazyLoadImage with EnhancedImage
const VirtualizedChampionsIndex = ({ onChampionSelect }) => {
  const [champions, setChampions] = useState([]);
  const [searchTerm, setSearchTerm] = useState('');
  
  useEffect(() => {
    // Your existing champion loading logic
    fetch('https://ddragon.leagueoflegends.com/cdn/14.9.1/data/en_US/champion.json')
      .then(response => response.json())
      .then(data => {
        const champList = Object.values(data.data);
        setChampions(champList);
      })
      .catch(error => console.error('Error fetching champions:', error));
  }, []);

  const filteredChampions = champions.filter(champion => 
    champion.name.toLowerCase().includes(searchTerm.toLowerCase())
  );

  // Calculate grid dimensions
  const ITEM_WIDTH = 120;
  const ITEM_HEIGHT = 150;
  const ITEMS_PER_ROW = 5; // Adjust based on your layout

  // Render a champion cell
  const ChampionCell = ({ champion, style }) => (
    <motion.div
      whileHover={{ scale: 1.05 }}
      whileTap={{ scale: 0.95 }}
      onClick={() => onChampionSelect(champion.id)}
      className="flex flex-col items-center p-2 cursor-pointer"
      style={style}
    >
      <div className="w-16 h-16 rounded-full overflow-hidden mb-2">
        <LazyLoadImage
          src={`https://ddragon.leagueoflegends.com/cdn/14.9.1/img/champion/${champion.image.full}`}
          alt={champion.name}
          effect="blur"
          width={64}
          height={64}
          className="w-full h-full object-cover"
          threshold={100}
          placeholderSrc="https://via.placeholder.com/64?text=..."
        />
      </div>
      <Text size="1" className="text-center text-gray-200">{champion.name}</Text>
    </motion.div>
  );

  // Cell renderer for react-window
  const Cell = ({ columnIndex, rowIndex, style }) => {
    const index = rowIndex * ITEMS_PER_ROW + columnIndex;
    if (index >= filteredChampions.length) return null;
    
    const champion = filteredChampions[index];
    return (
      <motion.div
        whileHover={{ scale: 1.05 }}
        whileTap={{ scale: 0.95 }}
        onClick={() => onChampionSelect(champion.id)}
        className="flex flex-col items-center p-2 cursor-pointer"
        style={style}
      >
        <div className="w-16 h-16 rounded-full overflow-hidden mb-2">
          <EnhancedImage
            src={`https://ddragon.leagueoflegends.com/cdn/14.9.1/img/champion/${champion.image.full}`}
            alt={champion.name}
            className="w-full h-full object-cover"
            style={{ backgroundColor: '#1e293b' }}
            onError={(e) => {
              console.error("Failed to load champion image:", champion.name);
            }}
          />
        </div>
        <Text size="1" className="text-center text-gray-200">{champion.name}</Text>
      </motion.div>
    );
  };

  return (
    <div className="w-full h-full">
      <div className="mb-4 px-4">
        <input
          type="text"
          placeholder="Search champions..."
          value={searchTerm}
          onChange={(e) => setSearchTerm(e.target.value)}
          className="w-full p-2 bg-gray-800 text-white rounded"
        />
      </div>
      
      <div className="w-full h-[calc(100%-60px)]">
        <AutoSizer>
          {({ height, width }) => {
            const columnCount = Math.floor(width / ITEM_WIDTH) || ITEMS_PER_ROW;
            const rowCount = Math.ceil(filteredChampions.length / columnCount);
            
            return (
              <FixedSizeGrid
                columnCount={columnCount}
                columnWidth={ITEM_WIDTH}
                height={height}
                rowCount={rowCount}
                rowHeight={ITEM_HEIGHT}
                width={width}
              >
                {Cell}
              </FixedSizeGrid>
            );
          }}
        </AutoSizer>
      </div>
    </div>
  );
};

// Similarly, update the SkinLinesIndex component
const VirtualizedSkinLinesIndex = ({ onSkinLineSelect }) => {
  const [skinLines, setSkinLines] = useState([]);
  const [searchTerm, setSearchTerm] = useState('');
  
  useEffect(() => {
    // Your existing skin lines loading logic
    fetch('https://raw.communitydragon.org/latest/plugins/rcp-be-lol-game-data/global/default/v1/skinlines.json')
      .then(response => response.json())
      .then(data => {
        const filteredLines = data.filter(line => line.id > 0);
        setSkinLines(filteredLines);
      })
      .catch(error => console.error('Error fetching skin lines:', error));
  }, []);

  const filteredSkinLines = skinLines.filter(line => 
    line.name.toLowerCase().includes(searchTerm.toLowerCase())
  );

  // Calculate grid dimensions
  const ITEM_WIDTH = 160;
  const ITEM_HEIGHT = 200;
  const ITEMS_PER_ROW = 4; // Adjust based on your layout

  // Render a skin line cell
  const SkinLineCell = ({ skinLine, style }) => (
    <motion.div
      whileHover={{ scale: 1.05 }}
      whileTap={{ scale: 0.95 }}
      onClick={() => onSkinLineSelect(skinLine.id)}
      className="flex flex-col items-center p-2 cursor-pointer"
      style={style}
    >
      <div className="w-24 h-24 rounded-lg overflow-hidden mb-2">
        <LazyLoadImage
          src={`https://raw.communitydragon.org/latest/plugins/rcp-be-lol-game-data/global/default/assets/skinlines/${skinLine.id.toString().padStart(4, '0')}-${skinLine.name.toLowerCase().replace(/\s+/g, '')}/splash.jpg`}
          alt={skinLine.name}
          effect="blur"
          width={96}
          height={96}
          className="w-full h-full object-cover"
          threshold={100}
          placeholderSrc="https://via.placeholder.com/96?text=..."
          onError={(e) => {
            e.target.onerror = null;
            e.target.src = "https://via.placeholder.com/96?text=No+Image";
          }}
        />
      </div>
      <Text size="1" className="text-center text-gray-200">{skinLine.name}</Text>
    </motion.div>
  );

  // Cell renderer for react-window
  const Cell = ({ columnIndex, rowIndex, style }) => {
    const index = rowIndex * ITEMS_PER_ROW + columnIndex;
    if (index >= filteredSkinLines.length) return null;
    
    const skinLine = filteredSkinLines[index];
    return <SkinLineCell skinLine={skinLine} style={style} />;
  };

  return (
    <div className="w-full h-full">
      <div className="mb-4 px-4">
        <input
          type="text"
          placeholder="Search skin lines..."
          value={searchTerm}
          onChange={(e) => setSearchTerm(e.target.value)}
          className="w-full p-2 bg-gray-800 text-white rounded"
        />
      </div>
      
      <div className="w-full h-[calc(100%-60px)]">
        <AutoSizer>
          {({ height, width }) => {
            const columnCount = Math.floor(width / ITEM_WIDTH) || ITEMS_PER_ROW;
            const rowCount = Math.ceil(filteredSkinLines.length / columnCount);
            
            return (
              <FixedSizeGrid
                columnCount={columnCount}
                columnWidth={ITEM_WIDTH}
                height={height}
                rowCount={rowCount}
                rowHeight={ITEM_HEIGHT}
                width={width}
              >
                {Cell}
              </FixedSizeGrid>
            );
          }}
        </AutoSizer>
      </div>
    </div>
  );
};

// Update the MainContent component to use the virtualized components
const MainContent = ({ activeTab, selectedChampion, setSelectedChampion, selectedSkinLine, setSelectedSkinLine }) => {
  const { champion } = useParams();
  const { userData, setUserData } = useUser();
  const { revalidateUser } = useUser();

  const [isLoading, setIsLoading] = useState(true);
  const handleChampionSelect = (championKey) => {
    setSelectedChampion(championKey);
  };
  const handleSkinLineSelect = (skinLineId) => {
    setSelectedSkinLine(skinLineId);
  };
  useEffect(() => {
    const fetchUserData = async () => {
      const token = localStorage.getItem("token");

      if (!token) {
        setIsLoading(false);
        return;
      }

      try {
        const response = await GetUserData(token);
        if (response.success) {
          setUserData(response.user);
        } else {
          throw new Error(response.error || 'Failed to load user data');
        }
      } catch (error) {
        console.error("Error fetching user data:", error);
        if (error.message.includes('token')) {
          localStorage.removeItem("token");
          setUserData(null);
        }
      } finally {
        setIsLoading(false);
      }
    };

    fetchUserData();
  }, [setUserData]);

  const handleLogout = () => {
    localStorage.removeItem("token");
    setUserData(null);
  };
  switch (activeTab) {

    case 'champions':
      if (selectedSkinLine) {
        return (
          <SkinLine skinLineId={selectedSkinLine} onBack={() => setSelectedSkinLine(null)} />
        );
      }
      return selectedChampion ? (
        <Champion champion={selectedChampion} onBack={() => setSelectedChampion(null)} />
      ) : (
        <ChampionsIndex onChampionSelect={handleChampionSelect} />
      );

    case 'installed':
      return (
        <Section className="bg-transparent">
          <Heading as="h1" size="5" align="center">Installed Skins</Heading>
          <ScrollArea
            className="mt-4"
            type="auto"
            scrollbars="vertical"
          >
            <InstalledSkinsDialog />
          </ScrollArea>
        </Section>
      );
    case 'profile':
      return (
        <Section className="pt-24 bg-transparent">
          <Container size="4" p="4" className="bg-transparent">
            <ProfileDialog userData={userData} onLogout={handleLogout} />
          </Container>
        </Section>
      );
    default:
      return <ChampionsIndex />;
  }
};

const AppInterface = () => {
  const [activeTab, setActiveTab] = useState('champions');
  const [isOmnisearchOpen, setIsOmnisearchOpen] = useState(false);
  const [currentStatus, setCurrentStatus] = useState('Waiting...');
  const isAuthenticated = !!localStorage.getItem("token");
  const [loginOpen, setLoginOpen] = useState(false);
  const [profileOpen, setProfileOpen] = useState(false);
  const [currentPopup, setCurrentPopup] = useState("login");
  const [selectedChampion, setSelectedChampion] = useState(null);
  const [selectedSkinLine, setSelectedSkinLine] = useState(null);
  const [showingSkinLines, setShowingSkinLines] = useState(false);
  const handleSkinLinesClick = () => {
    setShowingSkinLines(!showingSkinLines);
    setSelectedChampion(null);
    setSelectedSkinLine(null);
  };

  useEffect(() => {
    // Cargar el estado desde localStorage al iniciar
    const savedStatus = localStorage.getItem('modStatus');
    if (savedStatus) {
      setCurrentStatus(savedStatus);
    }

    // Create a global function to update status from anywhere
    window.updateGlobalStatus = (newStatus) => {
      console.log('New status received:', newStatus);
      
      // Clean up the status message if needed
      const formattedStatus = newStatus.replace("Status: ", "");
      
      setCurrentStatus(formattedStatus);
      
      // Save to localStorage for persistence
      localStorage.setItem('modStatus', formattedStatus);
    };

    // Listen for events from the backend
    EventsOn("overlay-stdout-update", (data) => {
      window.updateGlobalStatus(data.content);
    });
    
    EventsOn("overlay-stderr-update", (data) => {
      window.updateGlobalStatus(`Error: ${data.content}`);
    });

    return () => {
      // Clean up event listeners
      delete window.updateGlobalStatus;
      EventsOff("overlay-stdout-update");
      EventsOff("overlay-stderr-update");
    };
  }, []);

  return (
    <div className="flex flex-col h-screen text-gray-200">
      {/* Header */}
      <header className="bg-[#0f1729] w-full border-b-2 border-[#586a9e] fixed inline-flex items-center px-[13px] z-[71]" style={{ WebkitAppRegion: "drag", paddingTop: "1.1rem", paddingBottom: "1.1rem" }}>
        <img className="ml-[7px] w-52 aspect-[900/167]" src="https://i.imgur.com/m40l0qA.png" alt="cancer" />
        <Flex direction="row" className='absolute left-1/2 transform -translate-x-1/2' style={{ padding: "inherit" }}><Code variant='outline'>{currentStatus}</Code></Flex>
      </header>

      {/* Main content area with background image and transparent asides */}
      <section className={`flex flex-1 overflow-hidden z-50 ${activeTab === 'installed' ? 'pb-0' : 'pb-[154px]'}`}>
        {/* Background image layer that spans the entire width */}
        <div 
          className="absolute top-[73px] left-0 right-0 bottom-0 z-10" 
          style={{ 
            backgroundImage: "url('https://skinhunter.s-ul.eu/NJCK7P1q')",
          }}
        ></div>
        
        {/* Left Aside - now with transparent background */}
        <aside className="w-20 flex flex-col items-center py-4 h-dvh z-[60] backdrop-blur-sm relative">
          {activeTab === 'champions' && (selectedChampion || selectedSkinLine) ? (
            <motion.button
              whileHover={{ scale: 1.1 }}
              whileTap={{ scale: 0.95 }}
              onClick={() => {
                setSelectedChampion(null);
                setSelectedSkinLine(null);
              }}
              className="text-gray-400 hover:text-gray-200 mt-20 z-[61]"
            >
              <ArrowLeft size={24} />
            </motion.button>
          ) : activeTab === 'champions' && !selectedChampion && !selectedSkinLine && (
            <motion.button
              whileHover={{ scale: 1.1 }}
              whileTap={{ scale: 0.95 }}
              onClick={handleSkinLinesClick}
              className={`text-gray-400 hover:text-gray-200 mt-20 z-[61] ${showingSkinLines ? 'text-gray-200' : ''}`}
            >
              <AnimatePresence mode="wait">
                <IconToggle showingSkinLines={showingSkinLines} />
              </AnimatePresence>
            </motion.button>
          )}
        </aside>

        {/* Main Content */}
        <ScrollArea 
          type='scroll' 
          scrollbars="vertical" 
          className="flex-1 bg-transparent h-full mt-[73px] px-4 relative z-20" 
          style={{ height: "maxHeight", transform: "scaleY(1.03)" }}
        >
          <AnimatePresence mode="wait">
            {activeTab === 'installed' ? (
              <TabTransition key="installed">
                <div className="p-4">
                  <InstalledSkinsDialog />
                </div>
              </TabTransition>
            ) : activeTab === 'champions' && showingSkinLines && !selectedChampion && !selectedSkinLine ? (
              <TabTransition key="skinlines">
                <SkinLinesIndex onSkinLineSelect={(id) => setSelectedSkinLine(id)} />
              </TabTransition>
            ) : activeTab === 'champions' ? (
              <TabTransition key="champions">
                <MainContent
                  activeTab={activeTab}
                  selectedChampion={selectedChampion}
                  setSelectedChampion={setSelectedChampion}
                  selectedSkinLine={selectedSkinLine}
                  setSelectedSkinLine={setSelectedSkinLine}
                />
              </TabTransition>
            ) : activeTab === 'profile' && isAuthenticated ? (
              <TabTransition key="profile">
                <div className="p-4">
                  <ProfileDialog />
                </div>
              </TabTransition>
            ) : null}
          </AnimatePresence>
        </ScrollArea>

        {/* Right Aside - now with transparent background */}
        <aside className="w-20 flex flex-col items-center py-4 h-dvh z-[60]  backdrop-blur-sm relative">
          <ModOverlayButton></ModOverlayButton>
        </aside>
      </section>

      {/* Omnisearch Dialog */}
      <Omnisearch 
        isOpen={isOmnisearchOpen} 
        onOpenChange={setIsOmnisearchOpen}
        setActiveTab={setActiveTab}
        setSelectedChampion={setSelectedChampion}
        setSelectedSkinLine={setSelectedSkinLine}
      />

      {/* Bottom navigation */}
      <nav className="flex justify-center gap-16 py-4 bg-[#0f1729] border-t-2 border-[#586a9e] z-50 fixed w-full bottom-0 ">
        <button
          className={`flex flex-col items-center gap-1 transition-colors group ${activeTab === 'champions' ? 'text-gray-200' : 'text-gray-400 hover:text-gray-200'
            }`}
          onClick={() => setActiveTab('champions')}
        >
          <Grid className="w-5 h-5" />
          <span className="text-xs font-medium">Champions</span>
        </button>

        <button
          className={`flex flex-col items-center gap-1 transition-colors group ${activeTab === 'search' ? 'text-gray-200' : 'text-gray-400 hover:text-gray-200'
            }`}
          onClick={() => setIsOmnisearchOpen(true)}
        >
          <Search className="w-5 h-5" />
          <span className="text-xs font-medium">Search</span>
        </button>

        <button
          className={`flex flex-col items-center gap-1 transition-colors group ${activeTab === 'installed' ? 'text-gray-200' : 'text-gray-400 hover:text-gray-200'
            }`}
          onClick={() => setActiveTab('installed')}
        >
          <Download className="w-5 h-5" />
          <span className="text-xs font-medium">Installed</span>
        </button>

        {isAuthenticated && (
          <button
            className={`flex flex-col items-center gap-1 transition-colors group ${activeTab === 'settings' ? 'text-gray-200' : 'text-gray-400 hover:text-gray-200'
              }`}
            onClick={() => setActiveTab('profile')}
          >
            <User className="w-5 h-5" />
            <span className="text-xs font-medium">Profile</span>
          </button>
        )}

        {!isAuthenticated && (
          <Dialog.Root open={loginOpen} onOpenChange={setLoginOpen}>
            <Dialog.Trigger>
              <button
                className={`flex flex-col items-center gap-1 transition-colors group ${activeTab === 'settings' ? 'text-gray-200' : 'text-gray-400 hover:text-gray-200'
                  }`}
                onClick={() => open = { loginOpen }}
              >
                <User className="w-5 h-5" />
                <span className="text-xs font-medium">Profile</span>
              </button>
            </Dialog.Trigger>
            <Dialog.Content aria-describedby={undefined}>
              <ContainerForm
                closePopup={() => setLoginOpen(false)}
                setCurrentPopup={setCurrentPopup}
                currentPopup={currentPopup}
              />
            </Dialog.Content>
          </Dialog.Root>
        )}
      </nav>
    </div>
  );
};

export default AppInterface;