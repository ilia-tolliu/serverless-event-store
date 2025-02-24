import * as cdk from 'aws-cdk-lib';
import {aws_dynamodb, aws_events, aws_events_targets, aws_sns, CfnOutput} from 'aws-cdk-lib';
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

// import * as sqs from 'aws-cdk-lib/aws-sqs';

export class AwsEventStoreStack extends cdk.Stack {
    constructor(scope: Construct, id: string, props?: cdk.StackProps) {
        super(scope, id, props);

        const esTable = this.makeDynamoDbTable()

        const esLambda = this.makeLambdaFunction()

        const esUrl = this.addLambdaFunctionUrl(esLambda);

        const snsTopic = this.addNotifications(esTable)

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

    private makeLambdaFunction() {
        const esServiceRole = new Role(this, 'EsServiceRole', {
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
            loggingFormat: LoggingFormat.JSON
        })
    }

    private addLambdaFunctionUrl(fn: Function) {
        return fn.addFunctionUrl({
            authType: FunctionUrlAuthType.NONE // todo: use AWS_IAM auth type
        })
    }

    private addNotifications(esTable: TableV2) {
        const esTopic = new aws_sns.Topic(this, 'EsTopic', {
        })

        const cdcRule = new aws_events.Rule(this, 'EsChangesRule', {
            eventPattern: {
                source: ['aws.dynamodb'],
                detail: {
                    dynamodb: {
                        Keys: {
                            SK: {
                                N: ["0"]
                            }
                        }
                    },
                    eventSourceArn: [esTable.tableStreamArn]
                },
            }
        })

        cdcRule.addTarget(new aws_events_targets.SnsTopic(esTopic))

        return esTopic
    }

    private addSsmParameters(esTable: TableV2) {
        new StringParameter(this, 'DevelopmentEsTableName', {
            parameterName: '/development/event-store/DYNAMODB_TABLE_NAME',
            stringValue: esTable.tableName,
        });

        new StringParameter(this, 'DevelopmentEsPort', {
            parameterName: '/development/event-store/PORT',
            stringValue: '8080',
        });

        new StringParameter(this, 'StagingEsTableName', {
            parameterName: '/staging/event-store/DYNAMODB_TABLE_NAME',
            stringValue: esTable.tableName,
        });

        new StringParameter(this, 'StagingEsPort', {
            parameterName: '/staging/event-store/PORT',
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
