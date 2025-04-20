export const champions = [];
export const skinlines = [];
export const skins = {};
export let v = "";
import { FetchChampionJson } from "../../wailsjs/go/main/App";
const root = `https://raw.communitydragon.org/latest`,
  dataRoot = `${root}/plugins/rcp-be-lol-game-data/global/default`,
  dataRootFe = `${root}/plugins/rcp-fe-lol-champion-details`;




export const _champData = async (champId = null) => {


  if (champId) {

    const champDataUrl = `${dataRoot}/v1/champions/${champId}.json`;
    try {
      const response = await fetch(champDataUrl);
      if (!response.ok) {
        throw new Error(`Error al obtener datos del campeón: ${response.statusText}`);
      }
      return await response.json();
    } catch (error) {
      console.error(error.message);
      return null;
    }
  }


};



// export const _champDataName = async (champId = null) => {
//   if (!champId) return null;

//   try {
//     const basePath = window.electron.getChampionsPath(); // Obtiene la ruta absoluta
//     const response = await fetch(`file://${basePath}/${champId}.json`);

//     if (!response.ok) throw new Error("Not found");

//     return await response.json();
//   } catch (error) {
//     console.error(`Error: Champion Not found "${champId}".`, error);
//     return null;
//   }
// };



export const _champDataName = async (champId = null) => {
  if (!champId) return null;

  try {
    const response = await FetchChampionJson(String(champId));
    
    if (!response || !response.success) {
      throw new Error(`Error: Champion Not found "${champId}".`);
    }
    
    return response.data;
  } catch (error) {
    console.error(`Error: Champion Not found "${champId}".`, error);
    return null;
  }
};




export const _champChromas = async (champId = null) => {


  if (champId) {
    const proxyUrl = "https://cors-anywhere.herokuapp.com/";

    const champDataUrl = `https://cdn.merakianalytics.com/riot/lol/resources/latest/en-US/champions/${champId}.json`;
    try {
      const response = await fetch(proxyUrl + champDataUrl)

      if (!response.ok) {
        throw new Error(`Error al obtener datos del campeón: ${response.statusText}`);
      }
      return await response.json();
    } catch (error) {
      console.error(error.message);
      return null;
    }
  }


};


export const _ready = (async () => {
  const version = await fetch(`${root}/content-metadata.json`, {
    method: "GET",
    cache: "no-cache",
  }).then((r) => r.json());
  v = version.version;
  const cacheBreak = `?${encodeURIComponent(version.version)}`;

  await Promise.all([
    fetch(`${dataRoot}/v1/champion-summary.json${cacheBreak}`)
      .then((res) => res.json())
      .then((data) =>
        data
          .filter((d) => d.id !== -1)
          .sort((a, b) => (a.name > b.name ? 1 : -1))
          .map((a) => ({ ...a, key: a.alias.toLowerCase() }))
      )
      .then((data) => champions.push(...data)),
    fetch(`${dataRoot}/v1/skinlines.json${cacheBreak}`)
      .then((res) => res.json())
      .then((data) =>
        data
          .filter((d) => d.id !== 0)
          .sort((a, b) => (a.name > b.name ? 1 : -1))
      )
      .then((data) => skinlines.push(...data)),
    fetch(`${dataRoot}/v1/skins.json${cacheBreak}`)
      .then((res) => res.json())
      .then((data) => Object.assign(skins, data)),
  ]).then(() => true);
  return true;
})();

export function splitId(id) {
  return [Math.floor(id / 1000), id % 1000];
}

export function championSkins(id) {
  return Object.values(skins).filter((skin) => splitId(skin.id)[0] === id);
}

export function getChromasForSkin(skin) {
  if (!skin) {
    return null;
  }
  if (!skin.chromas || !Array.isArray(skin.chromas) || skin.chromas.length === 0) {
    return null;
  }
  return skin.chromas
    .filter(chroma => chroma && chroma.name && chroma.id)
    .map(chroma => ({
      origin: skin.id,
      name: chroma.name,
      id: chroma.id,
      chromaPath: chroma.chromaPath ? asset(chroma.chromaPath) : null,
      colors: chroma.colors || []
    }));
}


export function skinlineSkins(id) {
  return Object.values(skins)
    .filter((skin) => skin.skinLines?.some((line) => line.id === id))
    .sort((a, b) => {
      const aId = splitId(a.id)[0];
      const bId = splitId(b.id)[0];
      const aIndex = champions.findIndex((c) => c.id === aId);
      const bIndex = champions.findIndex((c) => c.id === bId);
      return aIndex - bIndex;
    });
}

export function asset(path) {
  return path.replace("/lol-game-data/assets", dataRoot).toLowerCase();
}

const rarities = {
  kUltimate: ["ultimate.png", "Ultimate"],
  kMythic: ["mythic.png", "Mythic"],
  kLegendary: ["legendary.png", "Legendary"],
  kEpic: ["epic.png", "Epic"],
  kTranscendent: ["transcendent.png", "Transcendent"],
  kExalted: ["exalted.png", "exalted"],

};

const damages = {
  kPhysic: ["kPhysic.png", "Physic"],
  kMixed: ["kMixed.png", "Mixed"],
  kMagic: ["kMagic.png", "Magic"],
}

const styles = {

  kASGrey: ["continuum_icon_attackspeed_grey.png", "Attack"],
  kASHighlight: ["continuum_icon_attackspeed.png", "Attack"],
  kAPGrey: ["continuum_icon_abilitypower_grey.png", "Magic"],
  kMagicHighlight: ["continuum_icon_abilitypower.png", "Magic"],
}

const legacyConfig = {
  legacyItems: ["icon-legacy.png", "Legacy"]
};

export function checkLegacy(skin) {
  if (Boolean(skin?.isLegacy)) {
    return [
      `https://raw.communitydragon.org/latest/plugins/rcp-fe-lol-static-assets/global/default/images/summoner-icon/${legacyConfig.legacyItems[0]}`,
      legacyConfig.legacyItems[1]
    ];
  }
  return null;
}

export function debugLegacySkins(skins) {
  const skinsInfo = skins.map(skin => ({
    name: skin.name,
    isLegacy: Boolean(skin.isLegacy)
  }));
  console.log("Skins cargadas:", skinsInfo);
}

export function rarity(skin) {
  if (!rarities[skin.rarity]) return null;
  const [imgName, name] = rarities[skin.rarity];
  const imgUrl = `${dataRoot}/v1/rarity-gem-icons/${imgName}`;
  return [imgUrl, name];
}

export function damageType(type) {
  if (!damages[type.damage_type]) return null;
  const [imgName, name] = damages[type.damage_type];
  const imgUrl = `${dataRootFe}/v1/rarity-gem-icons/${imgName}`;
  return [imgUrl, name];
}

export function style(s) {
  if (!styles[s.tipo]) {
    return ["default-icon.png", "Unknown"];
  }
  const [imgName, name] = styles[s.tipo];
  const imgUrl = `${dataRootFe}/global/default/${imgName}`;
  return [imgUrl, name];
}
// global/default/

export function KhadaUrl(skin, chroma) {
  if (chroma && skin !== chroma) {
    return `https://modelviewer.lol/model-viewer?id=${skin}&chroma=${chroma}`;
  }
  return `https://modelviewer.lol/model-viewer?id=${skin}`;
}
