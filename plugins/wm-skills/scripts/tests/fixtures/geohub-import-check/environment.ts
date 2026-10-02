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

export const shards: Shards = {
  local: {
    origin: 'http://127.0.0.1:8008/',
    elasticApi: 'http://localhost:9200/api/v2/search',
    graphhopperHost: 'https://graphhopper.webmapp.it/',
    awsApi: 'http://localhost:9080/wmfe/osm2cai2',
  },
  geohub: {
    origin: 'https://geohub.webmapp.it',
    elasticApi: 'https://elastic-json.webmapp.it/v2/search',
    graphhopperHost: 'https://graphhopper.webmapp.it/',
    awsApi: 'https://wmfe.s3.eu-central-1.amazonaws.com/geohub',
  },
  geohubdev: {
    origin: 'https://geohub.dev.maphub.it',
    elasticApi: 'http://geohub.dev.maphub.it:9200/v2/search',
    graphhopperHost: 'https://graphhopper.webmapp.it/',
    awsApi: 'https://geohub.dev.maphub.it/wmfe/geohub',
  },
  geohub2: {
    origin: 'https://geohub2.maphub.it',
    elasticApi: 'https://geohub2.maphub.it/api/v2/elasticsearch',
    graphhopperHost: 'https://graphhopper.webmapp.it/',
    awsApi: 'https://wmfe.s3.eu-central-1.amazonaws.com/geohub2',
  },
  osm2cai: {
    origin: 'https://osm2cai.cai.it',
    elasticApi: 'https://osm2cai.cai.it/api/v2/elasticsearch',
    graphhopperHost: 'https://graphhopper.webmapp.it/',
    awsApi: 'https://wmfe.s3.eu-central-1.amazonaws.com/osm2cai2',
  },
  osm2caiprod: {
    origin: 'https://osm2cai.prod.maphub.it',
    elasticApi: 'https://osm2cai.prod.maphub.it/api/v2/elasticsearch',
    graphhopperHost: 'https://graphhopper.webmapp.it/',
    awsApi: 'https://r3-it.storage.cloud.it/wmfe/osm2cai2',
  },
  osm2caiuat: {
    origin: 'https://osm2cai.uat.maphub.it',
    elasticApi: 'https://osm2cai.uat.maphub.it/api/v2/elasticsearch',
    graphhopperHost: 'https://graphhopper.webmapp.it/',
    awsApi: 'https://wmfe.s3.eu-central-1.amazonaws.com/osm2cai2uat',
  },
  osm2caidev: {
    origin: 'https://osm2cai.dev.maphub.it',
    elasticApi: 'https://osm2cai.dev.maphub.it/api/v2/elasticsearch',
    graphhopperHost: 'https://graphhopper.webmapp.it/',
    awsApi: 'https://osm2cai.dev.maphub.it/wmfe/osm2cai2dev',
  },
  osm2cailocal: {
    origin: 'http://127.0.0.1:8000/',
    elasticApi: 'http://localhost:9200/api/v2/search',
    graphhopperHost: 'https://graphhopper.webmapp.it/',
    awsApi: 'http://localhost:9002/wmfe/osm2cai2',
  },
  camminiditalia: {
    origin: 'https://camminiditalia.maphub.it',
    elasticApi: 'https://camminiditalia.maphub.it/api/v2/elasticsearch',
    graphhopperHost: 'https://graphhopper.webmapp.it/',
    awsApi: 'https://wmfe.s3.eu-central-1.amazonaws.com/camminiditalia',
  },
  camminiditaliadev: {
    origin: 'https://camminiditalia.dev.maphub.it',
    elasticApi: 'https://camminiditalia.dev.maphub.it/api/v2/elasticsearch',
    graphhopperHost: 'https://graphhopper.webmapp.it/',
    awsApi: 'https://camminiditalia.dev.maphub.it/wmfe/camminiditaliadev',
  },
  carg: {
    origin: 'https://carg.geosciences-ir.it',
    elasticApi: 'https://carg.geosciences-ir.it/api/v2/elasticsearch',
    graphhopperHost: 'https://graphhopper.webmapp.it/',
    awsApi: 'https://carg.geosciences-ir.it/storage/wmfe/carg',
  },
  cargdev: {
    origin: 'https://carg.maphub.it',
    elasticApi: 'https://carg.maphub.it/api/v2/elasticsearch',
    graphhopperHost: 'https://graphhopper.webmapp.it/',
    awsApi: 'https://wmfe.s3.eu-central-1.amazonaws.com/carg',
  },
  ersafdev: {
    origin: 'https://ersaf.dev.maphub.it',
    elasticApi: 'https://ersaf.dev.maphub.it/api/v2/elasticsearch',
    graphhopperHost: 'https://graphhopper.webmapp.it/',
    awsApi: 'https://ersaf.dev.maphub.it/wmfe/ersafdev',
  },
  forestas: {
    origin: 'https://forestas.maphub.it',
    elasticApi: 'https://forestas.maphub.it/api/v2/elasticsearch',
    graphhopperHost: 'https://graphhopper.webmapp.it/',
    awsApi: 'https://forestas.maphub.it/wmfe/forestas',
  },
  forestasdev: {
    origin: 'https://forestas.dev.maphub.it',
    elasticApi: 'https://forestas.dev.maphub.it/api/v2/elasticsearch',
    graphhopperHost: 'https://graphhopper.webmapp.it/',
    awsApi: 'https://forestas.dev.maphub.it/wmfe/forestasdev',
  },
  forestasuat: {
    origin: 'https://forestas.uat.maphub.it',
    elasticApi: 'https://forestas.uat.maphub.it/api/v2/elasticsearch',
    graphhopperHost: 'https://graphhopper.webmapp.it/',
    awsApi: 'https://forestas.uat.maphub.it/wmfe/forestasuat',
  },
  maphub: {
    origin: 'https://maphub.it',
    elasticApi: 'https://maphub.it/api/v2/elasticsearch',
    graphhopperHost: 'https://graphhopper.webmapp.it/',
    awsApi: 'https://wmfe.s3.eu-central-1.amazonaws.com/maphub',
  },
  maphubdev: {
    origin: 'https://dev.maphub.it',
    elasticApi: 'https://dev.maphub.it/api/v2/elasticsearch',
    graphhopperHost: 'https://graphhopper.webmapp.it/',
    awsApi: 'https://dev.maphub.it/wmfe/maphubdev',
  },
};

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
