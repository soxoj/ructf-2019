#include <stdio.h>
#include <stdlib.h>
#include <memory.h>

const unsigned int MEMORY_SIZE = 4;

void evil_func() {
    printf("Here is!\n");
}

struct Stack {
    int* stack;
    unsigned int cap;
    unsigned int len;
};

struct Stack* new_stack(unsigned int start_cap) {
    struct Stack* stack = malloc(sizeof(struct Stack));
    stack->stack = (int*)malloc(start_cap * sizeof(int));
    stack->cap = start_cap;
    stack->len = 0;
    return stack;
}

void push(struct Stack* stack, int x) {
    if (stack->len + 1 > stack->cap) {
        stack->stack = realloc(stack->stack, stack->cap * 2);
    }
    stack->stack[stack->len++] = x;
}

char pop(struct Stack* stack) {
    if (!stack->len) {
        return 1;
    }
    stack->len--;
    return 0;
}

char top(struct Stack* stack, int* x) {
    if (!stack->len) {
        return 1;
    }
    *x = stack->stack[stack->len - 1];
    return 0;
}

void free_stack(struct Stack* stack) {
    free(stack->stack);
    stack->stack = NULL;
    stack->len = 0;
    stack->cap = 0;
}

int skip(char* code, size_t code_size, int index) {
    int deep = 0;
    for (int i = index; i < code_size; ++i) {
        if (code[i] == '[') {
            deep++;
        } else if (code[i] == ']') {
            deep--;
        }
        if (!deep) {
            return i;
        }
    }
    return 0;
}

int run_bf_code(char* code, int code_len, char* input, int input_len, char* output, int max_output_len, unsigned int max_operations, unsigned int* written_bytes) {
    int output_pointer = 0;
    int input_pointer = 0;
    char cells[MEMORY_SIZE];
    int operations = 0;
    memset(cells, 0, MEMORY_SIZE);
    size_t pointer = 0;
    struct Stack* braces = new_stack(10);

    for (int i = 0; i < code_len; ++i) {
        if (code[i] == '+') {
            cells[pointer]++;
        } else if (code[i] == '-') {
            cells[pointer]--;
        } else if (code[i] == '.') {
            if (output_pointer >= max_output_len) {
                return 1;
            }
            output[output_pointer++] = cells[pointer];
        } else if (code[i] == ',') {
            if (input_pointer >= input_len) {
                return 2;
            }
            cells[pointer] = input[input_pointer++];
        } else if (code[i] == '>') {
            pointer++;
//            if (++pointer == MEMORY_SIZE) {
//                pointer = 0;
//            }
        } else if (code[i] == '<') {
            pointer--;
//            if (--pointer == -1) {
//                pointer = MEMORY_SIZE - 1;
//            }
        } else if (code[i] == '[') {

            if (!cells[pointer]) {
                int new_index = skip(code, code_len, i);
                if (!new_index) {
                    return 3;
                } else {
                    i = new_index;
                }
            } else {
                push(braces, i);
            }
        } else if (code[i] == ']') {
            if (!braces->len) {
                return 4;
            } else {
                top(braces, &i);
                i--;
                pop(braces);
            }
        }
        if (operations++ > max_operations) {
            return 5;
        }
    }
    if (braces->len) {
        return 6;
    }
    free_stack(braces);
    *written_bytes = output_pointer;
    return 0;
}
