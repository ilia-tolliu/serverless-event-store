import * as cdk from 'aws-cdk-lib';
import {aws_dynamodb, aws_sns, CfnOutput} from 'aws-cdk-lib';
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
import {Topic} from "aws-cdk-lib/aws-sns";
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

        const snsTopic = this.addNotifications(esTable, esLogs)

        this.addSsmParameters(esTable)

        this.makeStackOutputs(esTable, esLambda, esUrl, snsTopic)
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
                        'CreatedAt'
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
        const esTopic = new aws_sns.Topic(this, 'EsTopic', {})

        const pipeRole = new Role(this, 'EsPipeRole', {
            assumedBy: new ServicePrincipal('pipes.amazonaws.com'),
        })
        // pipeRole.addManagedPolicy('')
        esTable.grantStreamRead(pipeRole)
        esLogs.grantWrite(pipeRole)
        esTopic.grantPublish(pipeRole)

        new CfnPipe(this, 'EsPipe', {
            roleArn: pipeRole.roleArn,
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

    private addSsmParameters(esTable: TableV2) {
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
    }

    private makeStackOutputs(esTable: TableV2, esLambda: Function, esUrl: FunctionUrl, esTopic: Topic,) {
        new CfnOutput(this, 'OutputEsDynamoDbTable', {key: 'EsDynamoDbTable', value: esTable.tableName})
        new CfnOutput(this, 'OutputEsSnsTopic', {key: 'EsSnsTopic', value: esTopic.topicArn})
        new CfnOutput(this, 'OutputEsLogGroup', {key: 'EsLogGroup', value: esLambda.logGroup.logGroupName})
        new CfnOutput(this, 'OutputEsUrl', {key: 'EsUrl', value: esUrl.url})
    }
}
