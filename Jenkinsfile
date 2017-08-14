// pipeline script to build fhid and push build to s3 bucket
// depends on credentials being set up in Jenkins ahead of time
// along with the Dockerfile being built ahead of time as well.

def majVersion = "1"
def minVersion = "0"
def relVersion = "1"

def productName = "fhid"
def version = "${majVersion}.${minVersion}.${relVersion}.${env.BUILD_NUMBER}"
def packageNameNix = "${productName}-linux-amd64-${version}.tar.gz"
def packageNameNixLatest = "${productName}-linux-amd64-latest.tar.gz"
def bucketPath = "builds-fhid/"
def outfile = "./build/fhid"
def dkrWorkdir = "/go/src/github.build.ge.com/212601587/fhid"
def stageECRepo = "${productName}-stage"
def prodECRepo = "${productName}-prod"
def configFile = "stage-config.json"
def dkrImageNameStage = "${productName}-stage"
def dkrImageNameProd = "${productName}-prod"
def dkrImageNameBuilder = "${productName}-builder"
def stageTaskDefinitionFile = "ecs-task-stage.json"
def clusterName = "COPS-tainer-cluster"
def stageServiceName = "${productName}-service-stage"
def stageUrl = "https://images.cloudpod.apps.ge.com/v0.0/healthcheck"
def ecsControlNode = "docker-agent"
def slackChannel = "#cloudpod-feed-dev"
def jenkinsSecretBucketName = "s3-bucket-general-COPS-builds"


try {
    node("master"){
        withCredentials([
            string(credentialsId: "cloudpod-slack-token", variable: "SLACKTOKEN"),
                         string(credentialsId: "cloudpod-slack-org", variable: "SLACKORG"),
                         string(credentialsId: "${jenkinsSecretBucketName}", variable: "S3BUCKET")]) 
        {
            stage("cleanup") {
                deleteDir()
            }
            stage ("checkout source") {
                checkout scm
                sh "echo workspace path = ${env.WORKSPACE}"
            }
            stage ("build the build docker") {
                dir("./pipeline_build") {
                    sh "docker build . -t ${dkrImageNameBuilder}"
                }
            }
            stage ("test and build code in docker") {
                sh "docker run --rm -e GOOS=linux -e GOARCH=amd64 -v \"${env.WORKSPACE}\":${dkrWorkdir}:Z ${dkrImageNameBuilder} ${dkrWorkdir}/build.sh ${outfile} ${version} ${dkrWorkdir}"
            }
            stage ("package") {
                sh "cd ./build && tar zcfv ../${packageNameNix} . && cd .."
            }
            stage ("artifact upload") {
                awsIdentity()
                sh "/usr/bin/aws s3 cp ${packageNameNix} s3://${S3BUCKET}/${bucketPath}${packageNameNix}"
                sh "/usr/bin/aws s3 cp ${packageNameNix} s3://${S3BUCKET}/${bucketPath}${packageNameNixLatest}"
            }
            try {
                node(ecsControlNode) {
                    withCredentials([
                            string(credentialsId: "cloudpod-slack-token", variable: "SLACKTOKEN"),
                            string(credentialsId: "cloudpod-slack-org", variable: "SLACKORG"),
                            string(credentialsId: "${jenkinsSecretBucketName}", variable: "S3BUCKET"),
                            [$class: "UsernamePasswordMultiBinding", credentialsId: "dpco-s3-bucket-grabber",
                                usernameVariable: "ACCESS_KEY_ID", passwordVariable: "SECRET_ACCESS_KEY"]]) {
                        def pullCommand = "python pull-binary.py ${S3BUCKET} ${bucketPath}${packageNameNixLatest} ${ACCESS_KEY_ID} ${SECRET_ACCESS_KEY}"
                        stage("cleanup") {
                            deleteDir()
                        }
                        stage ("ecs-control-node: checkout source") {
                            checkout scm
                        }
                        stage ("ecs-control-node: pull latest binary") {
                            // requires boto3 on the jenkins agent
                            dir ("./pipeline_runtime") {
                                sh "${pullCommand}"
                            }
                        }
                        stage ("ecs-control-node: build runtime docker") {
                            dir ("./pipeline_runtime") {
                                sh "docker build -e CONFIG_FILE=${configFile} . -t ${dkrImageNameStage}"
                            }
                        }
                        stage ("ecs-control-node: push runtime container to stage repo") {
                            dir ("./pipeline_runtime") {
                                sh "python ecr-pusher.py ${stageECRepo} ${dkrImageNameStage} ${version}" 
                            }
                        }
                        stage ("ecs-control-node: kill existing staging env") {
                            dir ("./pipeline_runtime") {
                                sh "python ecs-killer.py ${clusterName} ${stageServiceName}" 
                            }
                        }
                        stage ("ecs-control-node: create stage task, start task, then create service") {
                            dir ("./pipeline_runtime") {
                                sh "python ecs-tasker-servicer.py ${stageTaskDefinitionFile} ${clusterName}" 
                            }
                        }
                        stage ("notify sucess") {
                            slackSend channel: "${slackChannel}", color: "good", message: "${productName} stage build and deployment SUCCEEDED. You can see version ${version} running here ${stageUrl}", teamDomain: "${SLACKORG}", token:"${SLACKTOKEN}"   
                        }
                    }
                }
            } catch (error) {
                node(ecsControlNode) {
                    echo "BUILD FAILED"
                    error "BUILD FAILED. See console logs"
                }
            } finally {
                node(ecsControlNode) {
                    stage("post build cleanup") {
                        sh "rm -rf ${env.WORKSPACE}/*"
                    }
                }
            }
        }
    }
} catch (error) {
    withCredentials([string(credentialsId: "cloudpod-slack-token", variable: "SLACKTOKEN"),
                     string(credentialsId: "cloudpod-slack-org", variable: "SLACKORG")])
    {
        stage ("notify failure") {
            slackSend channel: "${slackChannel}", color: "bad", message: "${productName} stage build FAILED ${env.BUILD_URL}", teamDomain: "${SLACKORG}", token:"${SLACKTOKEN}"   
        }
    }
} finally {
    node("master"){
        stage("post build cleanup") {
            sh "rm -rf ${env.WORKSPACE}/*"
        }
    }
}
