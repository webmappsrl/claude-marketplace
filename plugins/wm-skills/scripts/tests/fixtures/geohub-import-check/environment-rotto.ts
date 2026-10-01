export type Redirects = Readonly<Record<string, Redirect>>;

export type ShardName =
  | 'local'
  | 'geohub'
  | 'geohubdev'
  | 'geohub2'
  | 'osm2cai'
  | 'osm2caiprod'
  | 'osm2caiuat'
  | 'osm2caidev'
  | 'osm2cailocal'
  | 'camminiditalia'
  | 'camminiditaliadev'
  | 'carg'
  | 'cargdev'
  | 'ersafdev'
  | 'forestas'
  | 'forestasdev'
  | 'forestasuat'
  | 'maphub'
  | 'maphubdev';

export type Shards = Readonly<Record<ShardName, Shard>>;

export interface Environment {
  readonly production: boolean;
  readonly debug?: boolean;
  readonly appId: number;
  readonly shardName: ShardName;
  readonly shards: Shards;
  readonly redirects: Redirects;
}
export interface Shard {
  readonly origin: string;
  readonly elasticApi: string;
  readonly graphhopperHost: string;
  readonly awsApi: string;
}
export interface Redirect {
  readonly shardName: ShardName;
  readonly appId: number;
}


export const redirects: Redirects = {
  'sentieri.caiparma.it': {
    shardName: 'geohub',
    appId: 33,
  },
  'motomappa.motoabbigliamento.it': {
    shardName: 'geohub',
    appId: 53,
  },
  'maps.parcoforestecasentinesi.it': {
    shardName: 'geohub',
    appId: 49,
  },
  'maps.parcopan.org': {
    shardName: 'geohub',
    appId: 63,
  },
  'maps.acquasorgente.cai.it': {
    shardName: 'osm2cai',
    appId: 3,
  },
  'maps.caipontedera.it': {
    shardName: 'geohub',
    appId: 59,
  },
  'maps.parcapuane.it': {
    shardName: 'geohub',
    appId: 62,
  },
  'fiemaps.it': {
    shardName: 'geohub',
    appId: 29,
  },
  'fiemaps.eu': {
    shardName: 'geohub',
    appId: 29,
  },
  'maps.sentierodeiducati.it': {
    shardName: 'geohub',
    appId: 60,
  },
  'maps.valdicecinaoutdoor.it': {
    shardName: 'geohub',
    appId: 64,
  },
};
