import * as cdk from 'aws-cdk-lib';
import {aws_dynamodb, CfnOutput} from 'aws-cdk-lib';
import {Construct} from 'constructs';
import {ProjectionType, StreamViewType, TableV2} from "aws-cdk-lib/aws-dynamodb";
import {
    Architecture,
    Code,
    Function,
    FunctionUrl,
    FunctionUrlAuthType,
    LoggingFormat,
    Runtime
} from "aws-cdk-lib/aws-lambda";
import * as path from "node:path";
import {ManagedPolicy, Role, ServicePrincipal} from "aws-cdk-lib/aws-iam";
import {StringParameter} from 'aws-cdk-lib/aws-ssm';
import {LoggingProtocol, Topic} from "aws-cdk-lib/aws-sns";
import {CfnPipe, CfnPipeProps} from "aws-cdk-lib/aws-pipes"
import {LogGroup} from "aws-cdk-lib/aws-logs";
import {esConfig} from "./esConfig";

export class AwsEventStoreStack extends cdk.Stack {
    constructor(scope: Construct, id: string, props?: cdk.StackProps) {
        super(scope, id, props);

        const esTable = this.makeDynamoDbTable()

        const esLogs = new LogGroup(this, "EsLogs")

        const esLambda = this.makeLambdaFunction(esLogs)

        const esUrl = this.addLambdaFunctionUrl(esLambda);

        const esSnsTopic = this.addNotifications(esTable, esLogs)

        this.addSsmParameters(esTable, esUrl, esSnsTopic)

        this.makeStackOutputs(esTable, esLambda, esUrl, esSnsTopic)
    }

    private makeDynamoDbTable() {
        return new aws_dynamodb.TableV2(this, 'EsTable', {
            partitionKey: {
                name: 'PK',
                type: aws_dynamodb.AttributeType.STRING,
            },
            sortKey: {
                name: 'SK',
                type: aws_dynamodb.AttributeType.NUMBER,
            },
            removalPolicy: cdk.RemovalPolicy.RETAIN,
            globalSecondaryIndexes: [
                {
                    indexName: 'StreamIndex',
                    partitionKey: {
                        name: 'StreamType',
                        type: aws_dynamodb.AttributeType.STRING,
                    },
                    sortKey: {
                        name: 'UpdatedAt',
                        type: aws_dynamodb.AttributeType.STRING
                    },
                    projectionType: ProjectionType.INCLUDE,
                    nonKeyAttributes: [
                        'StreamRevision',
                    ]
                }
            ],
            dynamoStream: StreamViewType.NEW_IMAGE
        })
    }

    private makeLambdaFunction(esLogs: LogGroup) {
        const esServiceRole = new Role(this, 'EsLambdaRole', {
            assumedBy: new ServicePrincipal('lambda.amazonaws.com'),
        })

        esServiceRole.addManagedPolicy(ManagedPolicy.fromAwsManagedPolicyName('service-role/AWSLambdaBasicExecutionRole'))
        esServiceRole.addManagedPolicy(ManagedPolicy.fromAwsManagedPolicyName('AmazonSSMReadOnlyAccess'))
        esServiceRole.addManagedPolicy(ManagedPolicy.fromAwsManagedPolicyName('AmazonDynamoDBFullAccess'));

        return new Function(this, 'EsLambda', {
            runtime: Runtime.PROVIDED_AL2023,
            architecture: Architecture.ARM_64,
            handler: 'bootstrap',
            code: Code.fromAsset(path.join(__dirname, '../../../function.zip')),
            memorySize: 1024,
            role: esServiceRole,
            environment: {
                EVENT_STORE_MODE: 'staging'
            },
            loggingFormat: LoggingFormat.JSON,
            logGroup: esLogs
        })
    }

    private addLambdaFunctionUrl(fn: Function) {
        return fn.addFunctionUrl({
            authType: FunctionUrlAuthType.NONE // todo: use AWS_IAM auth type
        })
    }

    private addNotifications(esTable: TableV2, esLogs: LogGroup) {
        const deliveryLoggingRole = new Role(this, 'EsDeliveryLoggingRole', {
            assumedBy: new ServicePrincipal('sns.amazonaws.com'),
        })
        deliveryLoggingRole.addManagedPolicy(ManagedPolicy.fromAwsManagedPolicyName('service-role/AmazonSNSRole'))

        const esTopic = new Topic(this, 'EsTopic', {
            loggingConfigs: [
                {
                    protocol: LoggingProtocol.SQS,
                    successFeedbackSampleRate: 100,
                    successFeedbackRole: deliveryLoggingRole,
                    failureFeedbackRole: deliveryLoggingRole,
                }
            ]
        })

        const notificationsRole = new Role(this, 'EsNotificationRole', {
            assumedBy: new ServicePrincipal('pipes.amazonaws.com'),
        })
        esTable.grantStreamRead(notificationsRole)
        esLogs.grantWrite(notificationsRole)
        esTopic.grantPublish(notificationsRole)

        new CfnPipe(this, 'EsPipe', {
            roleArn: notificationsRole.roleArn,
            source: esTable.tableStreamArn!,
            sourceParameters: {
                dynamoDbStreamParameters: {
                    startingPosition: 'LATEST'
                },
                filterCriteria: {
                    filters: [{
                        pattern: `{ 
                            "dynamodb": { 
                                "NewImage": { 
                                    "RecordType": { 
                                        "S": ["stream"] 
                                    } 
                                } 
                            } 
                        }`
                    }]
                },
            },
            target: esTopic.topicArn,
            targetParameters: {
                inputTemplate: `{
                    "StreamId": <$.dynamodb.Keys.PK.S>,
                    "StreamType": <$.dynamodb.NewImage.StreamType.S>,
                    "StreamRevision": <$.dynamodb.NewImage.StreamRevision.N>
                }`
            },
            logConfiguration: {
                includeExecutionData: ['ALL'],
                level: 'ERROR',
                cloudwatchLogsLogDestination: {
                    logGroupArn: esLogs.logGroupArn,
                }
            },

        } as CfnPipeProps)

        return esTopic
    }

    private addSsmParameters(esTable: TableV2, esUrl: FunctionUrl, snsTopic: Topic) {
        const appMode = esConfig.appMode
        const prefix = appMode.charAt(0).toUpperCase() + appMode.slice(1)

        new StringParameter(this, `${prefix}EsTableName`, {
            parameterName: `/${appMode}/event-store/DYNAMODB_TABLE_NAME`,
            stringValue: esTable.tableName,
        });

        new StringParameter(this, `${prefix}EsPort`, {
            parameterName: `/${appMode}/event-store/PORT`,
            stringValue: '8080',
        });

        new StringParameter(this, `${prefix}EsUrl`, {
            parameterName: `/${appMode}/event-store/ES_URL`,
            stringValue: esUrl.url,
        })

        new StringParameter(this, `${prefix}EsSnsTopic`, {
            parameterName: `/${appMode}/event-store/ES_SNS_TOPIC`,
            stringValue: snsTopic.topicArn,
        })
    }

    private makeStackOutputs(esTable: TableV2, esLambda: Function, esUrl: FunctionUrl, esTopic: Topic,) {
        new CfnOutput(this, 'OutputEsDynamoDbTable', {key: 'EsDynamoDbTable', value: esTable.tableName})
        new CfnOutput(this, 'OutputEsSnsTopic', {key: 'EsSnsTopic', value: esTopic.topicArn})
        new CfnOutput(this, 'OutputEsLogGroup', {key: 'EsLogGroup', value: esLambda.logGroup.logGroupName})
        new CfnOutput(this, 'OutputEsUrl', {key: 'EsUrl', value: esUrl.url})
    }
}
