import { Controller, Logger, OnModuleDestroy, OnModuleInit } from "@nestjs/common";
import * as amqp from "amqplib";

import { SyncService } from "./sync.service";
import { RABBITMQ } from "../../messaging/rabbitmq/rabbitmq.constants";

@Controller()
export class SyncConsumer implements OnModuleInit, OnModuleDestroy {

  private readonly logger = new Logger(SyncConsumer.name);

  private connection: amqp.ChannelModel | null = null;
  private channel: amqp.Channel | null = null;


  constructor(
    private readonly syncService: SyncService,
  ) {}


  async onModuleInit(): Promise<void> {
    await this.connectAndConsume();
  }


  async onModuleDestroy(): Promise<void> {
    await this.channel?.close();
    await this.connection?.close();
  }


  private async connectAndConsume(): Promise<void> {

    this.connection = await amqp.connect(RABBITMQ.URL);
    this.channel = await this.connection.createChannel();




    await this.channel.consume(
      RABBITMQ.QUEUES.EMBEDDING_SYNC,
      async(message)=>{

this.logger.warn("=== SYNC CALLBACK ENTERED ===");

        if(!message){
          return;
        }


        try{

          const raw = message.content.toString();
this.logger.warn(raw);
          this.logger.log(
            `[SYNC][RAW] ${raw.substring(0,300)}`
          );


          const event = this.normalizeEvent(raw);


          if(!event){

            this.logger.error(
              "[SYNC] Invalid event"
            );

            this.channel?.nack(
              message,
              false,
              false
            );

            return;
          }


          this.logger.log(
            `[SYNC] RECEIVED point=${event.pointId}`
          );


          await this.syncService.process(event);


          this.logger.log(
            `[SYNC] SUCCESS point=${event.pointId}`
          );


          this.channel?.ack(message);


        }catch(error){

          this.logger.error(
            `[SYNC] FAILED ${
              error instanceof Error
              ? error.message
              : JSON.stringify(error)
            }`
          );


          this.channel?.nack(
            message,
            false,
            false
          );
        }

      },
      {
        noAck:false
      }
    );


    this.logger.log(
      `Sync consumer listening on ${RABBITMQ.QUEUES.EMBEDDING_SYNC}`
    );
  }



  private normalizeEvent(raw:string){

    try{

      const data = JSON.parse(raw);


      /*
       * embedding.generated event
       */
      if(
        data.eventId &&
        data.pointId &&
        data.vector
      ){

        return {

          event_id:data.eventId,

          event_type:
            RABBITMQ.ROUTING_KEYS.EMBEDDING_GENERATED,


          pointId:data.pointId,

          vector:data.vector,


          payload:data.payload ?? {},


          timestamp:
            data.timestamp ??
            new Date().toISOString()
        };

      }


      return data;


    }catch(error){

      this.logger.error(
        "Cannot parse sync event"
      );

      return null;
    }

  }

}