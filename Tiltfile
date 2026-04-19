# https://docs.tilt.dev/api.html#api.version_settings
version_settings(constraint='>=0.22.2')

# config.main_path is the absolute path to the Tiltfile being run
# https://docs.tilt.dev/api.html#modules.config.main_path
tiltfile_path = config.main_path

# https://github.com/bazelbuild/starlark/blob/master/spec.md#print
print("""
Starting 4ks Services
""".format(tiltfile=tiltfile_path))

# RESOURCES
k8s_yaml([
    'dev/deploy/api.yaml',
    'dev/deploy/web.yaml',
    'dev/deploy/fetcher.yaml',
    'dev/deploy/firestore.yaml',
    'dev/deploy/typesense.yaml',
    'dev/deploy/pubsub.yaml',
    # 'dev/deploy/jaeger.yaml'
])

# API
google_app_creds_path = 'dev/secrets/sbx-4ks-google-app-creds.json'
local(
    'if [ ! -f {path} ]; then echo "missing {path}; add the local Google service account JSON before starting api"; exit 1; fi'.format(path=google_app_creds_path),
    quiet=True
)
google_app_creds = str(read_file(google_app_creds_path))
google_app_creds_yaml = '\n'.join(['    ' + line for line in google_app_creds.split('\n')])
k8s_yaml(blob("""
apiVersion: v1
kind: Secret
metadata:
  name: api-google-app-creds
stringData:
  google-app-creds.json: |
{google_app_creds_yaml}
""".format(google_app_creds_yaml=google_app_creds_yaml)))
k8s_resource(objects=['api-google-app-creds'], new_name='api-google-app-creds', labels=['backend'])
k8s_resource(
    workload='api',
    port_forwards='0.0.0.0:5734:5000',
    labels=['backend'],
    resource_deps=['api-google-app-creds', 'pubsub', 'firestore', 'typesense']
)
docker_build(
    '4ks-api',
    context='.',
    dockerfile='apps/api/Dockerfile.dev',
    only=[
        'apps/api',
        'go.mod',
        'go.sum',
        'libs/go',
        'libs/reserved-words'
    ],
    live_update=[
        sync('apps/api/', '/code/apps/api'),
        sync('libs/go/', '/code/libs/go'),
        run(
            'go mod tidy',
            trigger=['apps/api/']
        )
    ]
)

# fetcher
k8s_resource(
    workload='fetcher',
    port_forwards='0.0.0.0:5889:5000',
    labels=['backend']
)
docker_build(
    'fetcher',
    context='.',
    dockerfile='apps/fetcher/Dockerfile.dev',
    only=[
        'apps/fetcher',
        'libs/go'
    ],
    live_update=[
        sync('apps/fetcher/', '/code'),
        run(
            'go mod download && go mod tidy',
            trigger=['apps/fetcher/go.mod', 'apps/fetcher/go.sum']
        )
    ]
)

# WEB
## package_json hack allows docker to cache npm install
local_resource('package_json', cmd='./apps/web/package_json.sh', deps=['pnpm-lock.yaml'])
k8s_resource(
    workload='web',
    port_forwards='0.0.0.0:5736:3000',
    labels=['web','next']
)
docker_build(
    'web',
    context='.',
    dockerfile='apps/web/Dockerfile.dev',
    only=[
        'apps/web',
        'libs/ts',
        'libs/reserved-words',
        'scripts',
        'PACKAGE_JSON',
        'package.json',
        'pnpm-lock.yaml',
        'pnpm-workspace.yaml',
        'tsconfig.base.json',
    ],
    live_update=[
        sync('libs/ts/api-fetch/dist', '/code/libs/ts/api-fetch/dist'),
        sync('apps/web', '/code/apps/web'),
        run(
            'pnpm install',
            trigger=['package.json', 'apps/web/package.json']
        )
    ]
)

# PUBSUB
k8s_resource('pubsub', port_forwards='8085:8085', labels=['database','pubsub'])
local_resource(
    name='init (pubsub)',
    cmd='./dev/scripts/init-pubsub.sh',
    resource_deps=['pubsub']
)

# DATA
k8s_resource('typesense', port_forwards='0.0.0.0:8108:8108', labels=['database','typesense'])
k8s_resource('firestore', port_forwards='8200:8200', labels=['database','firestore'])
local_resource(
    name='init (data)',
    cmd='./dev/scripts/init-data.sh',
    resource_deps=['firestore','typesense','api']
)

# OBSERVABILITY
# k8s_resource('jaeger', port_forwards=['9411:9411','5775:5775','6831:6831','6832:6832','5778:5778','16686:16686','14250:14250','14268:14268','14269:14269'], labels=['jaeger'])
