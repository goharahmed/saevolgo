#include <cstdio>
#include <cstdlib>
#include <cstring>
#include <fstream>
#include <string>

#include <unistd.h>
#include <sys/types.h> 
#include <sys/socket.h>
#include <netinet/in.h>
#include <netdb.h>
#include <pthread.h>

#define BUFSIZE 4096

int total_counter = 0;

struct arg_struct
{
	int num;
	std::string serveraddress;
	int portno;
	std::string logfile;
	int counter;
};

void *query(void* arguments)
{
	struct arg_struct *args = (struct arg_struct *)arguments;
	struct hostent *server = gethostbyname(args->serveraddress.c_str());
	
	int sockfd, n;
	struct sockaddr_in serv_addr;
	char buffer[BUFSIZE];
	
	sockfd = socket(AF_INET, SOCK_STREAM, 0);
	if (sockfd < 0)
	{
		std::printf("ERROR opening socket\n");
		return NULL;
	}
	
	bzero((char *) &serv_addr, sizeof(serv_addr));
	serv_addr.sin_family = AF_INET;
	bcopy((char *)server->h_addr, (char *)&serv_addr.sin_addr.s_addr, server->h_length);
	serv_addr.sin_port = htons(args->portno);
	if(connect(sockfd,(struct sockaddr *) &serv_addr,sizeof(serv_addr)) < 0)
	{
		std::printf("ERROR connecting\n");
		close(sockfd);
		return NULL;
	}

	n = read(sockfd,buffer,BUFSIZE);  // just clear the buffer
	
	bzero(buffer,BUFSIZE);
	std::sprintf(buffer, "auth ClueCon\n\n");
	n = write(sockfd,buffer,strlen(buffer));
	if(n < 0)
	{
		std::printf("error writing\n");
		close(sockfd);
		return NULL;
	}
	std::printf("authenticated! response:\n");
	bzero(buffer,BUFSIZE);
	n = read(sockfd,buffer,BUFSIZE);
	if(n < 0)
	{
		std::printf("error reading\n");
		close(sockfd);
		return NULL;
	}
	std::printf("%s\n\n",buffer);
	
	bzero(buffer,BUFSIZE);
	std::sprintf(buffer, "event plain ALL\n\n");
	n = write(sockfd,buffer,strlen(buffer));
	if(n < 0)
	{
		std::printf("error writing\n");
		close(sockfd);
		return NULL;
	}
	std::printf("requesting events! response:\n");
	bzero(buffer,BUFSIZE);
	n = read(sockfd,buffer,BUFSIZE);
	if(n < 0)
	{
		std::printf("error reading\n");
		close(sockfd);
		return NULL;
	}
	std::printf("%s\n\n",buffer);
	
	while(true)
	{
		n = read(sockfd,buffer,BUFSIZE);
		if(n < 0)
		{
			std::printf("error reading\n");
			close(sockfd);
			return NULL;
		}

		std::string buf = buffer;
		size_t pos = buf.find("HEARTBEAT");

		if(pos != std::string::npos)
		{
			total_counter++;
			args->counter++;
			if(args->counter == 3)
			{
				bzero(buffer,BUFSIZE);
				std::sprintf(buffer,"bgapi version\n\n");
				n = write(sockfd,buffer,strlen(buffer));
				if(n < 0)
				{
					std::printf("error writing version\n\n");
					close(sockfd);
					return NULL;
				}
       				bzero(buffer,BUFSIZE);
       				n = read(sockfd,buffer,BUFSIZE);
        			if(n < 0)
        			{
                			std::printf("error reading\n");
                			close(sockfd);
                			return NULL;
        			}
        			std::printf("%s\n\n",buffer);

				args->counter = 0;
			}
		}

	}
	
	bzero(buffer,BUFSIZE);
	std::sprintf(buffer, "exit\n\n");
	n = write(sockfd,buffer,strlen(buffer));
	if(n < 0)
	{
		std::printf("error writing\n");
		close(sockfd);
		return NULL;
	}
	std::printf("exiting! response:\n");
	bzero(buffer,BUFSIZE);
	n = read(sockfd,buffer,BUFSIZE);
	if(n < 0)
	{
		std::printf("error reading\n");
		close(sockfd);
		return NULL;
	}
	std::printf("%s\n\n",buffer);
	
	close(sockfd);

}

int main()
{
	pthread_t thread1, thread2;
	int  iret1, iret2;
    	struct arg_struct args1, args2;

	args1.counter = 0;	
	args1.serveraddress = "2.1.1.2";
	args1.portno = 8021;
	args1.logfile = "/var/www/html/logfile1";
	
	args2.counter = 0;	
	args2.serveraddress = "2.1.1.3";
	args2.portno = 8021;
	args2.logfile = "/var/www/html/logfile2";
	
	iret1 = pthread_create(&thread1, NULL, &query, (void*)&args1);
	if(iret1)
	{
		std::printf("Error - pthread_create() return code: %d\n",iret1);
		return -1;
	}

	iret2 = pthread_create(&thread2, NULL, &query, (void*)&args2);
	if(iret2)
	{
		std::printf("Error - pthread_create() return code: %d\n",iret2);
		return -1;
	}

	//pthread_join(thread1, NULL);
	//pthread_join(thread2, NULL);
	while(1)
	{
		printf("total: %d   server 1: %d   server 2: %d \n",total_counter,args1.counter,args2.counter);
		sleep(1);
	}		
	
	return 0;
}


