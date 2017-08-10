#include <stdio.h>
#include <stdlib.h>
#include <sys/ipc.h>

void usage(void);

int main(int argc, char *argv[])
{
        key_t key;
        if(argc != 3) usage();
        key = ftok(argv[1], argv[2][0]);
        printf("0x%x\n", key);
}

void usage(void)
{
        fprintf(stderr, "ftok - A utility for generating an ftok key\n");
        fprintf(stderr, "       This is a wrapper for ftok(3) of sys/ipc.h\n\n");
        fprintf(stderr, "Usage:   ftok <filepath> <id>\n");
        fprintf(stderr, "Example: ftok . a\n");
        exit(1);
}
