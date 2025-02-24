import * as cdk from 'aws-cdk-lib';
import {aws_dynamodb, aws_events, aws_events_targets, aws_sns, CfnOutput} from 'aws-cdk-lib';
import {Construct} from 'constructs';
import {ProjectionType, StreamViewType} from "aws-cdk-lib/aws-dynamodb";
import {Architecture, Code, Function, FunctionUrlAuthType, Runtime} from "aws-cdk-lib/aws-lambda";
import * as path from "node:path";
import {ManagedPolicy, Role, ServicePrincipal} from "aws-cdk-lib/aws-iam";
import {StringParameter} from 'aws-cdk-lib/aws-ssm';

// import * as sqs from 'aws-cdk-lib/aws-sqs';

export class AwsEventStoreStack extends cdk.Stack {
    constructor(scope: Construct, id: string, props?: cdk.StackProps) {
        super(scope, id, props);

        const esTable = new aws_dynamodb.TableV2(this, 'EsTable', {
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

        new StringParameter(this, 'EsTableName', {
            parameterName: '/development/event-store/DYNAMODB_TABLE_NAME',
            stringValue: esTable.tableName,
        });

        const esServiceRole = new Role(this, 'EsServiceRole', {
            assumedBy: new ServicePrincipal('lambda.amazonaws.com'),
        })
        esServiceRole.addManagedPolicy(ManagedPolicy.fromAwsManagedPolicyName('service-role/AWSLambdaBasicExecutionRole'))
        esServiceRole.addManagedPolicy(ManagedPolicy.fromAwsManagedPolicyName('AmazonSSMReadOnlyAccess'))

        const esLambda = new Function(this, 'EsLambda', {
            runtime: Runtime.PROVIDED_AL2023,
            architecture: Architecture.ARM_64,
            handler: 'bootstrap',
            code: Code.fromAsset(path.join(__dirname, '../../../function.zip')),
            memorySize: 1024,
            role: esServiceRole,
            environment: {
                EVENT_STORE_MODE: 'development'
            }
        })

        new StringParameter(this, 'EsPort', {
            parameterName: '/development/event-store/PORT',
            stringValue: '8080',
        });

        const esUrl = esLambda.addFunctionUrl({
            authType: FunctionUrlAuthType.NONE // todo: use AWS_IAM auth type
        })

        const snsTopic = new aws_sns.Topic(this, 'EsTopic', {
            topicName: 'EsTopic',
        })

        const cdcRule = new aws_events.Rule(this, 'EsChangesRule', {
            ruleName: 'EsChangeRule',
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

        cdcRule.addTarget(new aws_events_targets.SnsTopic(snsTopic))

        new CfnOutput(this, 'OutputEsDynamoDbTable', {key: 'EsDynamoDbTable', value: esTable.tableName})
        new CfnOutput(this, 'OutputEsSnsTopic', {key: 'EsSnsTopic', value: snsTopic.topicArn})
        new CfnOutput(this, 'OutputEsLogGroup', {key: 'EsLogGroup', value: esLambda.logGroup.logGroupName})
        new CfnOutput(this, 'OutputEsUrl', {key: 'EsUrl', value: esUrl.url})
    }
}
