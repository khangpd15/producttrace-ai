import {
  Injectable,
  OnModuleInit,
  OnModuleDestroy,
  Logger,
} from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import * as amqp from 'amqplib';

import { RABBITMQ } from './rabbitmq.constants';

import { UserRegisteredConsumer } from '../consumers/user-registered.consumer';
import { UserVerifiedConsumer } from '../consumers/user-verified.consumer';
import { PasswordResetConsumer } from '../consumers/password-reset.consumer';
import { ProductCreatedConsumer } from '../consumers/product-created.consumer';


@Injectable()
export class RabbitMQService
  implements OnModuleInit, OnModuleDestroy
{

  private readonly logger = new Logger(RabbitMQService.name);

  private connection: amqp.ChannelModel | null = null;
  private channel: amqp.Channel | null = null;

  private isConnecting = false;
  private isDestroyed = false;


  constructor(
    private readonly configService: ConfigService,

    private readonly userRegisteredConsumer: UserRegisteredConsumer,
    private readonly userVerifiedConsumer: UserVerifiedConsumer,
    private readonly passwordResetConsumer: PasswordResetConsumer,
    private readonly productCreatedConsumer: ProductCreatedConsumer,

  ) {}


  async onModuleInit() {
    await this.connect();
  }


  async onModuleDestroy() {
    this.isDestroyed = true;
    await this.close();
  }



  private async connect() {

    if (
      this.isConnecting ||
      this.connection ||
      this.isDestroyed
    ) {
      return;
    }


    this.isConnecting = true;


    const url =
      this.configService.get<string>('RABBITMQ_URL')
      || RABBITMQ.URL;


    const maskedUrl =
      url.replace(/:([^:@]+)@/, ':****@');


    this.logger.log(
      `Connecting RabbitMQ ${maskedUrl}`
    );


    try {

      this.connection =
        await amqp.connect(url);


      this.logger.log(
        'RabbitMQ connected'
      );


      this.connection.on(
        'error',
        err => {

          this.logger.error(
            'RabbitMQ connection error',
            err
          );

        }
      );


      this.connection.on(
        'close',
        () => {

          this.logger.warn(
            'RabbitMQ connection closed'
          );


          this.handleDisconnect();

        }
      );


      await this.createChannel();


      this.isConnecting = false;


    } catch(error) {


      this.logger.error(
        'RabbitMQ connection failed',
        error
      );


      await this.close();


      this.isConnecting = false;


      this.handleDisconnect();

    }

  }





  private async createChannel(){

    if(!this.connection)
      return;


    try {


      this.channel =
        await this.connection.createChannel();



      this.channel.on(
        'error',
        err => {

          this.logger.error(
            'RabbitMQ channel error',
            err
          );

        }
      );


      this.channel.on(
        'close',
        ()=>{

          this.logger.warn(
            'RabbitMQ channel closed'
          );

        }
      );



      await this.setupTopology();



    } catch(error){


      this.logger.error(
        'Create channel failed',
        error
      );


      throw error;

    }

  }







  private async setupTopology(){

    if(!this.channel)
      return;



    const ch = this.channel;



    /*
      MAIN EXCHANGE
    */

    await ch.assertExchange(
      RABBITMQ.EXCHANGE,
      RABBITMQ.EXCHANGE_TYPE,
      {
        durable:true
      }
    );



    /*
      DEAD LETTER EXCHANGE
    */

    await ch.assertExchange(
      RABBITMQ.DLX_EXCHANGE,
      'topic',
      {
        durable:true
      }
    );



    const topology = [

      {
        name:
          RABBITMQ.QUEUES.USER_REGISTERED,

        routingKey:
          RABBITMQ.ROUTING_KEYS.USER_REGISTERED,

        consumer:
          this.userRegisteredConsumer,

      },


      {
        name:
          RABBITMQ.QUEUES.USER_VERIFIED,

        routingKey:
          RABBITMQ.ROUTING_KEYS.USER_VERIFIED,

        consumer:
          this.userVerifiedConsumer,

      },


      {
        name:
          RABBITMQ.QUEUES.PASSWORD_RESET,

        routingKey:
          RABBITMQ.ROUTING_KEYS.PASSWORD_RESET,

        consumer:
          this.passwordResetConsumer,

      },


      {
        name:
          RABBITMQ.QUEUES.PRODUCT_CREATED,

        routingKey:
          RABBITMQ.ROUTING_KEYS.PRODUCT_CREATED,

        consumer:
          this.productCreatedConsumer,

      },

    ];



    for(const item of topology){



      const failedQueue =
        `${item.name}.failed`;



      /*
        FAILED QUEUE
      */

      await ch.assertQueue(
        failedQueue,
        {
          durable:true
        }
      );



      await ch.bindQueue(
        failedQueue,
        RABBITMQ.DLX_EXCHANGE,
        failedQueue
      );



      /*
        MAIN QUEUE
      */

      await ch.assertQueue(
        item.name,
        {

          durable:true,


          arguments:{

            'x-dead-letter-exchange':
              RABBITMQ.DLX_EXCHANGE,


            'x-dead-letter-routing-key':
              failedQueue

          }

        }
      );



      await ch.bindQueue(
        item.name,
        RABBITMQ.EXCHANGE,
        item.routingKey
      );



      this.logger.log(
        `Consumer started: ${item.name}`
      );



      await ch.consume(

        item.name,

        async msg=>{


          if(!msg)
            return;


          try{


            await item.consumer.handleMessage(
              msg,
              ch
            );


          }catch(error){


            this.logger.error(
              `Consumer error ${item.name}`,
              error
            );


          }


        },

        {
          noAck:false
        }

      );


    }


  }







  private handleDisconnect(){


    if(this.isDestroyed)
      return;



    this.connection=null;
    this.channel=null;



    setTimeout(()=>{

      this.connect();

    },5000);


  }







  private async close(){


    try{


      if(this.channel){

        await this.channel.close();

        this.channel=null;

      }



      if(this.connection){

        await this.connection.close();

        this.connection=null;

      }


    }catch(error){


      this.logger.error(
        'RabbitMQ close error',
        error
      );


    }

  }


}